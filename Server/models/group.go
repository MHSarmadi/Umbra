package models

import "time"

type Group struct {
	UUID                      []byte    `json:"uuid"`
	XPublicKey                []byte    `json:"x_pub_key"`
	EPublicKey                []byte    `json:"e_pub_key"`
	EncipheredEntranceKey     []byte    `json:"enciphered_entrance_key"`
	EncipheredEntranceKeySalt []byte    `json:"enciphered_entrance_key_salt"`
	EncipheredEntranceKeyTag  []byte    `json:"enciphered_entrance_key_tag"`
	CreatorUUID               []byte    `json:"creator_uuid"`
	CreatedAt                 time.Time `json:"created_at"`
}
