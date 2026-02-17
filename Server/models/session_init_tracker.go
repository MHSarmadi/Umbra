package models

import "time"

type SessionInitTracker struct {
	IdentityHash  string    `json:"identity_hash"`
	RequestUnixTS []int64   `json:"request_unix_ts"`
	ExpiresAt     time.Time `json:"expires_at"`
}

func (t *SessionInitTracker) Key() []byte {
	return append([]byte{0x12}, []byte(t.IdentityHash)...)
}
