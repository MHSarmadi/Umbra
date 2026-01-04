package models

import "time"

type Message struct {
	UUID            []byte    `json:"uuid"`
	GroupUUID       []byte    `json:"group_uuid"`
	XPublicKey      []byte    `json:"x_pub_key"`
	Payload         []byte    `json:"payload"`
	SenderUUID      []byte    `json:"sender_uuid"`
	SenderSignature []byte    `json:"sender_signature"`
	CreatedAt       time.Time `json:"created_at"`
}
