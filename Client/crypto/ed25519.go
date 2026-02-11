package crypto

import (
	"crypto/ed25519"
)

func DeriveEd25519PubKey(soul []byte) []byte {
	seed := KDF(soul, "@ED25519-PRIVATEKEY-DERIVATION", 32)
	return ed25519.NewKeyFromSeed(seed)[32:]
}

func Sign(soul, msg []byte) []byte {
	seed := KDF(soul, "@ED25519-PRIVATEKEY-DERIVATION", 32)
	return ed25519.Sign(ed25519.NewKeyFromSeed(seed), msg)
}

func Verify(pub, msg, sig []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(pub), msg, sig)
}
