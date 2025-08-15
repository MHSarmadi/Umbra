package crypto

import "github.com/zeebo/blake3"

func MAC(key, data []byte, context string) (digest [32]byte) {
	safeKey := make([]byte, 32)
	blake3.DeriveKey("@UMBRAv0.0.0-@STDMAC-"+context, key, safeKey)
	h, _ := blake3.NewKeyed(safeKey)
	h.Write(data)
	h.Digest().Read(digest[:])
	return
}
