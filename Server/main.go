package main

import (
	// "encoding/base64"
	"fmt"

	"crypto/rand"
	"math/bits"

	"github.com/MHSarmadi/Umbra/Server/crypto"
)

// func main() {
// s := crypto.Sum([]byte("Hello, World!"))
// fmt.Println(base64.RawURLEncoding.EncodeToString(s[:]))

// k := []byte("pass1234")
// m := crypto.MAC(k, []byte("Hello, World"), "TEST")
// fmt.Println(base64.RawURLEncoding.EncodeToString(m[:]))

// p := crypto.KDF(k, "TEST", 128)
// fmt.Println(base64.RawURLEncoding.EncodeToString(p))
// fmt.Println()

// // e, salt := crypto.MACE_Encrypt(k, []byte("Hello, World"), "TEST", 2, true)
// e, salt, tag := crypto.MACE_Encrypt_MIXIN_AEAD(k, []byte("Hello, World!"), []byte("I'm Authorized!"), "TEST", 2, false)
// fmt.Println(base64.RawURLEncoding.EncodeToString(e))
// fmt.Println(base64.RawURLEncoding.EncodeToString(salt))
// fmt.Println(base64.RawURLEncoding.EncodeToString(tag))

// // r, err := crypto.MACE_Decrypt(k, e, salt, "TEST", 2)
// // salt[0] ^= 1
// r, valid, err := crypto.MACE_Decrypt_MIXIN_AEAD(k, e, []byte("I'm Authorized!"), salt, tag, "TEST", 2)
// if err != nil {
// 	// panic(err)
// }
// fmt.Println(string(r))
// if valid {
// 	fmt.Println("VALID!")
// } else {
// 	fmt.Println("INVALID!!!")
// }
// fmt.Println("Length:", len(r))

// fmt.Println()
// fmt.Println()
// fmt.Println()

// msg := "RRRRENO"
// msg_buf := []byte(msg)
// msg_buf[0] ^= 0b1000000
// msg = string(msg_buf)
// key := "RRRRRRRRRENO"
// difficulty := 0

// res, salt := crypto.MACE_Encrypt([]byte(key), []byte(msg), "BARAYE KHANDE", uint16(difficulty), true)

// fmt.Println(base64.RawURLEncoding.EncodeToString(res))
// _ = salt

// chizi_ke_baz_shode, _ := crypto.MACE_Decrypt([]byte(key), res, salt, "BARAYE KHANDE", uint16(difficulty))
// fmt.Println(string(chizi_ke_baz_shode))

// }

// save as avalanche.go
// package main

// func hamming512(a, b []byte) int {
// 	count := 0
// 	for i := range 64 {
// 		count += bits.OnesCount8(a[i] ^ b[i])
// 	}
// 	return count
// }

// func H(seed []byte, diff int) (c []byte) {
// 	c, _ = crypto.MACE_Encrypt([]byte{0}, seed, "PRNG", uint16(diff), true)
// 	return
// }

//	func main() {
//		trials := 80000
//		for i := range 8 {
//			sum := 0.0
//			sumSq := 0.0
//			for range trials {
//				seed := make([]byte, 32)
//				_, err := rand.Read(seed)
//				if err != nil {
//					panic(err)
//				}
//				c := H(seed, i)
//				// flip the most-significant bit of seed[0]
//				seed2 := make([]byte, 32)
//				copy(seed2, seed)
//				seed2[0] ^= 0x80
//				c2 := H(seed2, i)
//				d := hamming512(c, c2)
//				sum += float64(d)
//				sumSq += float64(d * d)
//			}
//			mean := sum / float64(trials)
//			variance := sumSq/float64(trials) - mean*mean
//			// std := math.Sqrt(variance)
//			// fmt.Printf("trials=%d mean=%.4f variance=%.4f (expected mean=256, variance=128)\n", trials, mean, variance)
//			fmt.Printf("diff=%d mean=%.4f%% variance=%.4f%%\n", i, (mean/256-1)*100, (variance/128-1)*100)
//		}
//	}
func hamming512(a, b []byte) int {
	count := 0
	for i := range 64 {
		count += bits.OnesCount8(a[i] ^ b[i])
	}
	return count
}

// H is your deterministic PRNG/encryption core (MACE-BLAKE3).
func H(seed []byte, diff int) []byte {
	c, _ := crypto.MACE_Encrypt([]byte{0}, seed, "PRNG", uint16(diff), true)
	return c
}

func main() {
	trials := 100_000
	seedLen := 32 // bytes
	totalBits := seedLen * 8

	for diff := range 8 {
		sum := 0.0
		sumSq := 0.0

		for range trials {
			// generate random seed
			seed := make([]byte, seedLen)
			_, err := rand.Read(seed)
			if err != nil {
				panic(err)
			}

			// choose random bit position to flip (0..255)
			var flipBuf [1]byte
			_, _ = rand.Read(flipBuf[:])
			bitIndex := int(flipBuf[0]) % totalBits // random bit index
			byteIndex := bitIndex / 8
			bitMask := byte(1 << (7 - (bitIndex % 8))) // MSB first

			// compute c and c2
			c := H(seed, diff)

			seed2 := make([]byte, seedLen)
			copy(seed2, seed)
			seed2[byteIndex] ^= bitMask

			c2 := H(seed2, diff)

			// measure hamming distance
			d := hamming512(c, c2)
			sum += float64(d)
			sumSq += float64(d * d)
		}

		mean := sum / float64(trials)
		variance := sumSq/float64(trials) - mean*mean
		fmt.Printf(
			"diff=%d mean=%.4f%% variance=%.4f%%\n",
			diff,
			(mean/256-1)*100,
			(variance/128-1)*100,
		)
	}
}
