package models_requests

type SessionInitRequestEncoded struct {
	ClientEdPubKey         string `json:"client_ed_pubkey"`
	ClientXPubKey          string `json:"client_x_pubkey"`
	ClientXPubKeySignature string `json:"client_x_pubkey_sign"`
}

type SessionInitRequestDecoded struct {
	ClientEdPubKey         []byte
	ClientXPubKey          []byte
	ClientXPubKeySignature []byte
}
