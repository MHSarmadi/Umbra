package models

import "time"

type User struct {
	UUID               []byte    `json:"uuid"`
	Username           string    `json:"username"`
	XPublicKey         []byte    `json:"x_pub_key"`
	EPublicKey         []byte    `json:"e_pub_key"`
	EncipheredSoul     []byte    `json:"enciphered_soul"`
	EncipheredSoulSalt []byte    `json:"enciphered_soul_salt"`
	EncipheredSoulTag  []byte    `json:"enciphered_soul_tag"`
	SoulRecovery       []byte    `json:"soul_recovery"`
	SoulRecoverySalt   []byte    `json:"soul_recovery_salt"`
	SoulRecoveryTag    []byte    `json:"soul_recovery_tag"`
	CreatedAt          time.Time `json:"created_at"`
}

func (u *User) KeyByUUID() []byte {
	return append([]byte{0x10}, u.UUID...)
}
func (u *User) KeyByUsername() []byte {
	return append([]byte{0x11}, []byte(u.Username)...)
}
