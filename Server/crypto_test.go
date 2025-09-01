package main

import (
	"crypto/rand"
	"crypto/subtle"
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

func BenchmarkEncryption(b *testing.B) {
	buffer := make([]byte, 1024)
	key := []byte("Some Password")
	for i := range b.N {
		rand.Read(buffer)
		cipher, salt := crypto.MACE_Encrypt(key, buffer, "UNIT_TESTING"+string(rune(i)), 2, false)
		raw, _ := crypto.MACE_Decrypt(key, cipher, salt, "UNIT_TESTING"+string(rune(i)), 2)
		if subtle.ConstantTimeCompare(raw, buffer) != 1 {
			b.Fail()
		}
	}
}

func BenchmarkEncryptionMIXIN(b *testing.B) {
	buffer := make([]byte, 1024)
	key := []byte("Some Password")
	mixin := []byte("I'm Authorized!")
	for i := range b.N {
		rand.Read(buffer)
		cipher, salt := crypto.MACE_Encrypt_MIXIN(key, buffer, mixin, "UNIT_TESTING"+string(rune(i)), 2, false)
		raw, _ := crypto.MACE_Decrypt_MIXIN(key, cipher, mixin, salt, "UNIT_TESTING"+string(rune(i)), 2)
		if subtle.ConstantTimeCompare(raw, buffer) != 1 {
			b.Fail()
		}
	}
}

func BenchmarkEncryptionAEAD(b *testing.B) {
	buffer := make([]byte, 1024)
	key := []byte("Some Password")
	for i := range b.N {
		rand.Read(buffer)
		cipher, salt, tag := crypto.MACE_Encrypt_AEAD(key, buffer, "UNIT_TESTING"+string(rune(i)), 2, false)
		raw, _, _ := crypto.MACE_Decrypt_AEAD(key, cipher, salt, tag, "UNIT_TESTING"+string(rune(i)), 2)
		if subtle.ConstantTimeCompare(raw, buffer) != 1 {
			b.Fail()
		}
	}
}

func BenchmarkEncryptionMIXINAEAD(b *testing.B) {
	buffer := make([]byte, 1024)
	key := []byte("Some Password")
	mixin := []byte("I'm Authorized!")
	for i := range b.N {
		rand.Read(buffer)
		cipher, salt, tag := crypto.MACE_Encrypt_MIXIN_AEAD(key, buffer, mixin, "UNIT_TESTING"+string(rune(i)), 2, false)
		raw, _, _ := crypto.MACE_Decrypt_MIXIN_AEAD(key, cipher, mixin, salt, tag, "UNIT_TESTING"+string(rune(i)), 2)
		if subtle.ConstantTimeCompare(raw, buffer) != 1 {
			b.Fail()
		}
	}
}
