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

func coreEncrypt(hasher *blake3.Hasher, src, roundChunkBuf, latestChunk, digest []byte, rounds uint16, chunkSize byte) {
	var (
		chunk uint32
		i     byte
	)
	for round := range rounds {
		binary.BigEndian.PutUint16(roundChunkBuf[:2], round)
		copy(latestChunk, src[len(src)-int(chunkSize):])
		for chunk = uint32(len(src)/int(chunkSize) - 1); chunk > 0; chunk-- {
			binary.BigEndian.PutUint32(roundChunkBuf[2:6], chunk)
			hasher.Reset()
			if chunk >= 2 {
				hasher.Write(src[int(chunkSize)*int(chunk-2) : int(chunkSize)*int(chunk-1)])
			} else {
				hasher.Write(latestChunk)
			}
			hasher.Write(roundChunkBuf[:6])
			hasher.Digest().Read(digest)
			for i = range chunkSize {
				src[int(i)+int(chunkSize)*int(chunk)] = digest[i] ^ src[int(i)+int(chunkSize)*int(chunk-1)]
			}
		}
		copy(src[:chunkSize], latestChunk)
	}
}

func coreDecrypt(hasher *blake3.Hasher, src, roundChunkBuf, firstChunk, digest []byte, rounds uint16, chunkSize byte) {
	var (
		chunk uint32
		i     byte
	)
	for round := rounds - 1; round < rounds; round-- {
		binary.BigEndian.PutUint16(roundChunkBuf[:2], round)
		copy(firstChunk, src[:chunkSize])
		for chunk = uint32(0); chunk < uint32(len(src)/int(chunkSize))-1; chunk++ {
			binary.BigEndian.PutUint32(roundChunkBuf[2:6], chunk+1)
			hasher.Reset()
			if chunk >= 1 {
				hasher.Write(src[int(chunkSize)*int(chunk-1) : int(chunkSize)*int(chunk)])
			} else {
				hasher.Write(firstChunk)
			}
			hasher.Write(roundChunkBuf[:6])
			hasher.Digest().Read(digest)
			for i = range chunkSize {
				src[int(i)+int(chunkSize)*int(chunk)] = digest[i] ^ src[int(i)+int(chunkSize)*int(chunk+1)]
			}
		}
		copy(src[len(src)-int(chunkSize):], firstChunk)
	}
}

func internal_MACE_Encrypt(key, data []byte, context string, difficulty uint16, deterministic bool) (cipher, salt []byte, h *blake3.Hasher) {
	var chunkSize byte
	cipher = pkcs7Pad(data, 64)
	if len(cipher) == 64 {
		chunkSize = 32
	} else {
		chunkSize = 64
	}
	buffer := make([]byte, 50+2*chunkSize)
	safeKey := buffer[0:32]
	salt = buffer[38+2*chunkSize : 50+2*chunkSize]
	if !deterministic {
		if _, err := rand.Read(salt); err != nil {
			panic("crypto/rand failure: " + err.Error())
		}
	}
	h = blake3.NewDeriveKey("@UMBRAv0.0.0-@STDMACE-@MACEv1.0.0-" + context)
	h.Write(key)
	h.Write(salt)
	h.Digest().Read(safeKey)
	h, err := blake3.NewKeyed(safeKey)
	if err != nil {
		panic("blake3.NewKeyed failed: " + err.Error())
	}

	coreEncrypt(
		h,                                   // hasher
		cipher,                              // src
		buffer[32:38],                       // roundChunkBuf
		buffer[38:38+chunkSize],             // latestChunk
		buffer[38+chunkSize:38+2*chunkSize], // digest
		2*difficulty+3,                      // rounds
		chunkSize,                           // chunkSize
	)
	return
}

func internal_MACE_Decrypt(key, cipher, salt []byte, context string, difficulty uint16) (raw []byte, h *blake3.Hasher, err error) {
	var chunkSize byte
	if len(cipher) == 64 {
		chunkSize = 32
	} else {
		chunkSize = 64
	}
	buffer := make([]byte, 38+2*chunkSize)
	safeKey := buffer[0:32]
	h = blake3.NewDeriveKey("@UMBRAv0.0.0-@STDMACE-@MACEv1.0.0-" + context)
	h.Write(key)
	h.Write(salt)
	h.Digest().Read(safeKey)
	h, _ = blake3.NewKeyed(safeKey)
	coreDecrypt(
		h,                                   // hasher
		cipher,                              // src
		buffer[32:38],                       // roundChunkBuf
		buffer[38:38+chunkSize],             // firstChunk
		buffer[38+chunkSize:38+2*chunkSize], // digest
		2*difficulty+3,                      // rounds
		chunkSize,                           // chunkSize
	)
	raw, err = pkcs7Unpad(cipher, 64)
	return
}

func MACE_Encrypt(key, data []byte, context string, difficulty uint16, deterministic bool) (cipher, salt []byte) {
	cipher, salt, _ = internal_MACE_Encrypt(key, data, context, difficulty, deterministic)
	return
}

func MACE_Decrypt(key, cipher, salt []byte, context string, difficulty uint16) (raw []byte, err error) {
	if len(cipher)%64 != 0 || len(cipher) == 0 {
		return nil, errors.New("invalid input length - not correctly padded")
	}
	raw, _, err = internal_MACE_Decrypt(key, cipher, salt, context, difficulty)
	return
}

func MACE_Encrypt_MIXIN(key, data, mixin []byte, context string, difficulty uint16, deterministic bool) (cipher, salt []byte) {
	mixin_hash := blake3.Sum512(mixin)
	cipher, salt, _ = internal_MACE_Encrypt(append(key, mixin_hash[:]...), data, "@MIXIN-"+context, difficulty, deterministic)
	return
}

func MACE_Decrypt_MIXIN(key, cipher, mixin, salt []byte, context string, difficulty uint16) (raw []byte, err error) {
	if len(cipher)%64 != 0 || len(cipher) == 0 {
		return nil, errors.New("invalid input length - not correctly padded")
	}
	mixin_hash := blake3.Sum512(mixin)
	raw, _, err = internal_MACE_Decrypt(append(key, mixin_hash[:]...), cipher, salt, "@MIXIN-"+context, difficulty)
	return
}

func MACE_Encrypt_AEAD(key, data []byte, context string, difficulty uint16, deterministic bool) (cipher, salt, tag []byte) {
	tag = make([]byte, 16+2) // the latest 2 bytes is used to store difficulty in BigEndian
	binary.BigEndian.PutUint16(tag[16:], difficulty)
	cipher, salt, h := internal_MACE_Encrypt(key, data, "@AEAD-"+context, difficulty, deterministic)
	h.Reset()
	h.Write(cipher)
	h.Write(tag[16:])
	h.Digest().Read(tag)
	return
}

func MACE_Decrypt_AEAD(key, cipher, salt, tag []byte, context string, difficulty uint16) (raw []byte, valid bool, err error) {
	if len(cipher)%64 != 0 || len(cipher) == 0 {
		return nil, false, errors.New("invalid input length - not correctly padded")
	}
	expectedTag := make([]byte, len(cipher)+2) // also used to temp-store ciphered data // the latest 2 bytes is used to store difficulty in BigEndian
	binary.BigEndian.PutUint16(expectedTag[len(cipher):], difficulty)
	copy(expectedTag, cipher)
	raw, h, err := internal_MACE_Decrypt(key, cipher, salt, "@AEAD-"+context, difficulty)
	h.Reset()
	h.Write(expectedTag[:len(cipher)])
	h.Write(expectedTag[len(cipher):])
	h.Digest().Read(expectedTag[:16])
	valid = subtle.ConstantTimeCompare(tag, expectedTag[:16]) == 1
	return
}

func MACE_Encrypt_MIXIN_AEAD(key, data, mixin []byte, context string, difficulty uint16, deterministic bool) (cipher, salt, tag []byte) {
	tag = make([]byte, 16+2) // the latest 2 bytes is used to store difficulty in BigEndian
	binary.BigEndian.PutUint16(tag[16:], difficulty)
	mixin_hash := blake3.Sum512(mixin)
	cipher, salt, h := internal_MACE_Encrypt(append(key, mixin_hash[:]...), data, "@MIXIN-@AEAD-"+context, difficulty, deterministic)
	h.Reset()
	h.Write(cipher)
	h.Write(tag[16:])
	h.Write(mixin)
	h.Digest().Read(tag[:16])
	tag = tag[:16]
	return
}

func MACE_Decrypt_MIXIN_AEAD(key, cipher, mixin, salt, tag []byte, context string, difficulty uint16) (raw []byte, valid bool, err error) {
	if len(cipher)%64 != 0 || len(cipher) == 0 {
		return nil, false, errors.New("invalid input length - not correctly padded")
	}
	expectedTag := make([]byte, len(cipher)+2) // also used to temp-store ciphered data // the latest 2 bytes is used to store difficulty in BigEndian
	binary.BigEndian.PutUint16(expectedTag[len(cipher):], difficulty)
	mixin_hash := blake3.Sum512(mixin)
	copy(expectedTag[:len(cipher)], cipher)
	raw, h, err := internal_MACE_Decrypt(append(key, mixin_hash[:]...), cipher, salt, "@MIXIN-@AEAD-"+context, difficulty)
	h.Reset()
	h.Write(expectedTag[:len(cipher)])
	h.Write(expectedTag[len(cipher):])
	h.Write(mixin)
	h.Digest().Read(expectedTag[:16])
	valid = subtle.ConstantTimeCompare(tag, expectedTag[:16]) == 1
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
