package crypto

func DeriveX25519PubKey(soul []byte) ([]byte, error) {
	priv := KDF(soul, "@X25519-PRIVATEKEY-DERIVATION", 32)
	pub, err := X25519(priv, Basepoint)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

func ComputeSharedSecret(soul, peerPub []byte) ([]byte, error) {
	priv := KDF(soul, "@X25519-PRIVATEKEY-DERIVATION", 32)
	shared, err := X25519(priv, peerPub)
	if err != nil {
		return nil, err
	}
	return shared, nil
}