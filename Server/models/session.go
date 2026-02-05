package models

import "time"

type PowParamsType struct {
	MemoryMB    uint `json:"memory_mb"`
	Iterations  uint `json:"iterations"`
	Parallelism uint `json:"parallelism"`
}

type Session struct {
	UUID [24]byte `json:"uuid"`

	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`

	ClientEdPubKey [32]byte `json:"client_ed_pubkey"`
	ClientXPubKey  [32]byte `json:"client_x_pubkey"`

	ServerSoul [32]byte `json:"server_soul"`

	SessionToken [24]byte `json:"session_token"`

	LastNonces   map[string]int64 `json:"last_nonces"`   // int64: unix timestamp "seconds"
	LastActivity int64            `json:"last_activity"` // int64: unix timestamp "seconds"

	PoWChallenge    [2]byte       `json:"pow_challenge"`
	PoWParams       PowParamsType `json:"pow_params"`
	PoWSolution     []byte        `json:"pow_solution"`
}

func (u *Session) KeyByUUID() []byte {
	return append([]byte{0x10}, u.UUID[:]...)
}