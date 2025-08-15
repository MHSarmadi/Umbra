package main

import (
	"encoding/binary"
	"testing"

	"github.com/MHSarmadi/Umbra/Server/crypto"
)

func BenchmarkHash(b *testing.B) {
	buffer := make([]byte, 8)
	for i := range b.N {
		binary.BigEndian.PutUint64(buffer, uint64(i))
		_ = crypto.Sum(buffer)
	}
}

func BenchmarkKDF(b *testing.B) {
	buffer := make([]byte, 8)
	for i := range b.N {
		binary.BigEndian.PutUint64(buffer, uint64(i))
		_ = crypto.KDF(buffer, "UNIT_TESTING", 16)
	}
}

func BenchmarkMAC(b *testing.B) {
	buffer := make([]byte, 8)
	key := []byte("Random Password")
	for i := range b.N {
		binary.BigEndian.PutUint64(buffer, uint64(i))
		_ = crypto.MAC(key, buffer, "UNIT_TESTING")
	}
}
