package crypto

import "github.com/zeebo/blake3"

func Sum(data []byte) (digest [64]byte) {
	return blake3.Sum512(data)
}
