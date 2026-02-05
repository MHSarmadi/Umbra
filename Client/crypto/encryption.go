// ================================================================
// ⚠️  SECURITY & USAGE HAZARDS – READ BEFORE USING
// ================================================================
// This file implements the core symmetric encryption routines for Umbra.
// It intentionally deviates from common AEAD API safety conventions.
//
// 1) Decrypt functions **always return a plaintext** (dummy or real),
//    regardless of authentication/tag verification result. The caller
//    MUST check the `valid` boolean returned by AEAD variants before
//    using the plaintext. Failure to do so will result in the use of
//    unauthenticated, potentially attacker-controlled data.
//
//    Reason: This design avoids timing differences between valid and
//    invalid decryptions, but increases the risk of API misuse.
//
// 2) `*_MIXIN` variants do NOT provide authenticated associated data.
//    They only bind the MIXIN into the key derivation. If you need true
//    authenticated MIXIN, use the `*_MIXIN_AEAD` variants.
//
// 3) Deterministic mode (`deterministic = true`) disables salt
//    randomization and will produce identical ciphertext for identical
//    (key, context, plaintext[, MIXIN]) inputs. This leaks message
//    equality and must be used only in contexts where this is intended.
//
// 4) Decrypt and encrypt operations mutate the input buffers in place.
//    Copy your data first if you need to preserve the original.
//
// 5) Padding removal (`pkcs7Unpad`) is designed to be constant-time-ish
//    and will still return a truncated output even on padding errors.
//    Callers MUST check the returned `err` before using the output.
//
// Developers integrating this code into higher-level APIs must enforce
// correct usage patterns to avoid cryptographic misuse.
// ================================================================

package crypto

// BASED ON https://github.com/MHSarmadi/MACE

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"errors"

	"github.com/zeebo/blake3"
)

func clear(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func coreEncrypt(hasher *blake3.Hasher, mutSrc []byte, rounds uint16, chunkSize byte) {
	var (
		chunks              = uint32(len(mutSrc)/int(chunkSize) - 1)
		latestChunkPosition = len(mutSrc) - int(chunkSize)
		chunk               uint32
		i                   byte
		roundChunkBuf       [6]byte
		latestChunk         [64]byte
		digest              [64]byte
	)
	defer clear(digest[:])
	for round := range rounds {
		binary.BigEndian.PutUint16(roundChunkBuf[:], round)
		copy(latestChunk[:chunkSize], mutSrc[latestChunkPosition:])
		for chunk = chunks; chunk > 0; chunk-- {
			thisChunk := int(chunkSize) * int(chunk)
			prevChunk := thisChunk - int(chunkSize)
			binary.BigEndian.PutUint32(roundChunkBuf[2:], chunk)
			hasher.Reset()
			if chunk >= 2 {
				hasher.Write(mutSrc[prevChunk-int(chunkSize) : prevChunk])
			} else {
				hasher.Write(latestChunk[:chunkSize])
			}
			hasher.Write(roundChunkBuf[:])
			hasher.Digest().Read(digest[:chunkSize])
			for i = range chunkSize {
				mutSrc[thisChunk+int(i)] = digest[i] ^ mutSrc[prevChunk+int(i)]
			}
		}
		copy(mutSrc[:chunkSize], latestChunk[:chunkSize])
	}
}

func coreDecrypt(hasher *blake3.Hasher, mutSrc []byte, rounds uint16, chunkSize byte) {
	var (
		chunks              = uint32(len(mutSrc)/int(chunkSize)) - 1
		latestChunkPosition = len(mutSrc) - int(chunkSize)
		chunk               uint32
		i                   byte
		roundChunkBuf       [6]byte
		firstChunk          [64]byte
		digest              [64]byte
	)
	defer clear(digest[:])
	for round := rounds - 1; round < rounds; round-- {
		binary.BigEndian.PutUint16(roundChunkBuf[:], round)
		copy(firstChunk[:chunkSize], mutSrc[:chunkSize])
		for chunk = range chunks {
			thisChunk := int(chunkSize) * int(chunk)
			nextChunk := thisChunk + int(chunkSize)
			binary.BigEndian.PutUint32(roundChunkBuf[2:], chunk+1)
			hasher.Reset()
			if chunk >= 1 {
				hasher.Write(mutSrc[thisChunk-int(chunkSize) : thisChunk])
			} else {
				hasher.Write(firstChunk[:chunkSize])
			}
			hasher.Write(roundChunkBuf[:])
			hasher.Digest().Read(digest[:chunkSize])
			for i = range chunkSize {
				mutSrc[thisChunk+int(i)] = digest[i] ^ mutSrc[nextChunk+int(i)]
			}
		}
		copy(mutSrc[latestChunkPosition:], firstChunk[:chunkSize])
	}
}

func internal_MACE_Encrypt(key, data []byte, context string, difficulty uint16, deterministic bool) (cipher, salt []byte, h *blake3.Hasher) {
	var (
		chunkSize = byte(64)
		safeKey   [32]byte
		saltBuf   [12]byte
	)
	defer clear(safeKey[:])
	cipher = pkcs7Pad(data, 64)
	if len(cipher) == 64 {
		chunkSize = 32
	}
	salt = saltBuf[:]
	if !deterministic {
		if _, err := rand.Read(salt); err != nil {
			panic("crypto/rand failure: " + err.Error())
		}
	}
	h = blake3.NewDeriveKey("@UMBRAv0.0.0-@STDMACE-@MACEv1.0.0-" + context)
	h.Write(key)
	h.Write(salt)
	h.Digest().Read(safeKey[:])
	h, err := blake3.NewKeyed(safeKey[:])
	if err != nil {
		panic("blake3.NewKeyed failed: " + err.Error())
	}

	coreEncrypt(
		h,              // hasher
		cipher,         // src
		2*difficulty+3, // rounds
		chunkSize,      // chunkSize
	)
	return
}

func internal_MACE_Decrypt(key, mutCipher, salt []byte, context string, difficulty uint16) (raw []byte, h *blake3.Hasher, err error) {
	var (
		chunkSize = byte(64)
		safeKey   [32]byte
	)
	defer clear(safeKey[:])
	if len(mutCipher) == 64 {
		chunkSize = 32
	}
	h = blake3.NewDeriveKey("@UMBRAv0.0.0-@STDMACE-@MACEv1.0.0-" + context)
	h.Write(key)
	h.Write(salt)
	h.Digest().Read(safeKey[:])
	h, _ = blake3.NewKeyed(safeKey[:])
	coreDecrypt(
		h,              // hasher
		mutCipher,      // src
		2*difficulty+3, // rounds
		chunkSize,      // chunkSize
	)
	raw, err = pkcs7Unpad(mutCipher, 64)
	return
}

func MACE_Encrypt(key, data []byte, context string, difficulty uint16, deterministic bool) (cipher, salt []byte) {
	cipher, salt, _ = internal_MACE_Encrypt(key, data, context, difficulty, deterministic)
	return
}

func MACE_Decrypt(key, mutCipher, salt []byte, context string, difficulty uint16) (raw []byte, err error) {
	if len(mutCipher)%64 != 0 || len(mutCipher) == 0 {
		return nil, errors.New("invalid input length - not correctly padded")
	}
	raw, _, err = internal_MACE_Decrypt(key, mutCipher, salt, context, difficulty)
	return
}

func MACE_Encrypt_MIXIN(key, data, mixin []byte, context string, difficulty uint16, deterministic bool) (cipher, salt []byte) {
	mixin_hash := blake3.Sum512(mixin)
	cipher, salt, _ = internal_MACE_Encrypt(append(key, mixin_hash[:]...), data, "@MIXIN-"+context, difficulty, deterministic)
	return
}

func MACE_Decrypt_MIXIN(key, mutCipher, mixin, salt []byte, context string, difficulty uint16) (raw []byte, err error) {
	if len(mutCipher)%64 != 0 || len(mutCipher) == 0 {
		return nil, errors.New("invalid input length - not correctly padded")
	}
	mixin_hash := blake3.Sum512(mixin)
	raw, _, err = internal_MACE_Decrypt(append(key, mixin_hash[:]...), mutCipher, salt, "@MIXIN-"+context, difficulty)
	return
}

func MACE_Encrypt_AEAD(key, data []byte, context string, difficulty uint16, deterministic bool) (cipher, salt, tag []byte) {
	var (
		tagBuf        [16]byte
		difficultyBuf [2]byte
	)
	defer clear(tagBuf[:])
	binary.BigEndian.PutUint16(difficultyBuf[:], difficulty)
	cipher, salt, h := internal_MACE_Encrypt(key, data, "@AEAD-"+context, difficulty, deterministic)
	h.Reset()
	h.Write(cipher)
	h.Write(difficultyBuf[:])
	h.Digest().Read(tagBuf[:])
	tag = append([]byte(nil), tagBuf[:]...)
	return
}

func MACE_Decrypt_AEAD(key, cipher, salt, tag []byte, context string, difficulty uint16) (raw []byte, valid bool, err error) {
	if len(cipher)%64 != 0 || len(cipher) == 0 {
		return nil, false, errors.New("invalid input length - not correctly padded")
	}
	var (
		difficultyBuf [2]byte
		expectedTag   [16]byte
	)
	binary.BigEndian.PutUint16(difficultyBuf[:], difficulty)
	raw, h, err := internal_MACE_Decrypt(key, append([]byte(nil), cipher...), salt, "@AEAD-"+context, difficulty)
	h.Reset()
	h.Write(cipher)
	h.Write(difficultyBuf[:])
	h.Digest().Read(expectedTag[:])
	valid = subtle.ConstantTimeCompare(tag, expectedTag[:]) == 1
	return
}

func MACE_Encrypt_MIXIN_AEAD(key, data, mixin []byte, context string, difficulty uint16, deterministic bool) (cipher, salt, tag []byte) {
	var (
		tagBuf        [16]byte
		difficultyBuf [2]byte
	)
	defer clear(tagBuf[:])
	binary.BigEndian.PutUint16(difficultyBuf[:], difficulty)
	mixin_hash := blake3.Sum512(mixin)
	cipher, salt, h := internal_MACE_Encrypt(append(key, mixin_hash[:]...), data, "@MIXIN-@AEAD-"+context, difficulty, deterministic)
	h.Reset()
	h.Write(cipher)
	h.Write(difficultyBuf[:])
	h.Write(mixin)
	h.Digest().Read(tagBuf[:])
	tag = append([]byte(nil), tagBuf[:]...)
	return
}

func MACE_Decrypt_MIXIN_AEAD(key, cipher, mixin, salt, tag []byte, context string, difficulty uint16) (raw []byte, valid bool, err error) {
	if len(cipher)%64 != 0 || len(cipher) == 0 {
		return nil, false, errors.New("invalid input length - not correctly padded")
	}
	var (
		difficultyBuf [2]byte
		expectedTag   [16]byte
	)
	binary.BigEndian.PutUint16(difficultyBuf[:], difficulty)
	mixin_hash := blake3.Sum512(mixin)
	raw, h, err := internal_MACE_Decrypt(append(key, mixin_hash[:]...), append([]byte(nil), cipher...), salt, "@MIXIN-@AEAD-"+context, difficulty)
	h.Reset()
	h.Write(cipher)
	h.Write(difficultyBuf[:])
	h.Write(mixin)
	h.Digest().Read(expectedTag[:])
	valid = subtle.ConstantTimeCompare(tag, expectedTag[:]) == 1
	return
}

func pkcs7Pad(data []byte, blockSize byte) []byte {
	padding := int(blockSize) - (len(data) % int(blockSize))
	pad := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, pad...)
}

func pkcs7Unpad(data []byte, blockSize byte) (result []byte, err error) {
	if len(data) == 0 || len(data)%int(blockSize) != 0 {
		return nil, errors.New("invalid padded data length")
	}
	padLen := data[len(data)-1]
	if padLen == 0 || padLen > blockSize {
		err = errors.New("invalid padding")
		padLen = 1
	}
	for _, b := range data[len(data)-int(padLen):] {
		if b != padLen {
			err = errors.New("invalid padding")
			padLen = 1
		}
	}
	return data[:len(data)-int(padLen)], err
}
