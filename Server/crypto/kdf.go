package crypto

import "github.com/zeebo/blake3"

func KDF(rawKey []byte, context string, length uint16) (digest []byte) {
	digest = make([]byte, length)
	h := blake3.NewDeriveKey("@UMBRA-STDKDF-" + context)
	h.Write(rawKey)
	h.Digest().Read(digest)
	return
}
