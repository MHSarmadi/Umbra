package main

import (
	"encoding/base64"
	"fmt"

	"github.com/MHSarmadi/Umbra/Server/crypto"
)

func main() {
	s := crypto.Sum([]byte("Hello, World!"))
	fmt.Println(base64.RawURLEncoding.EncodeToString(s[:]))

	k := []byte("pass1234")
	m := crypto.MAC(k, []byte("Hello, World"), "TEST")
	fmt.Println(base64.RawURLEncoding.EncodeToString(m[:]))

	p := crypto.KDF(k, "TEST", 128)
	fmt.Println(base64.RawURLEncoding.EncodeToString(p))
	fmt.Println()

	// e, salt := crypto.MACE_Encrypt(k, []byte("Hello, World"), "TEST", 2, true)
	e, salt, tag := crypto.MACE_Encrypt_MIXIN_AEAD(k, []byte("Hello, World!"), []byte("I'm Authorized!"), "TEST", 2, false)
	fmt.Println(base64.RawURLEncoding.EncodeToString(e))
	fmt.Println(base64.RawURLEncoding.EncodeToString(salt))
	fmt.Println(base64.RawURLEncoding.EncodeToString(tag))

	// r, err := crypto.MACE_Decrypt(k, e, salt, "TEST", 2)
	// salt[0] ^= 1
	r, valid, err := crypto.MACE_Decrypt_MIXIN_AEAD(k, e, []byte("I'm Authorized!"), salt, tag, "TEST", 2)
	if err != nil {
		// panic(err)
	}
	fmt.Println(string(r))
	if valid {
		fmt.Println("VALID!")
	} else {
		fmt.Println("INVALID!!!")
	}
	fmt.Println("Length:", len(r))
}
