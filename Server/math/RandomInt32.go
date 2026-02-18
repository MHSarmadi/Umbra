package math_tools

import (
	"crypto/rand"
	"encoding/binary"
)

func RandomInt32(min, max int32) int32 {
	if min >= max {
		panic("invalid range: min must be < max")
	}

	var (
		rangeSize = uint32(max - min)
		maxUint   = ^uint32(0)
		limit     = maxUint - (maxUint % rangeSize)
	)

	for {
		var buf [4]byte
		if _, err := rand.Read(buf[:]); err != nil {
			panic(err)
		}

		r := binary.BigEndian.Uint32(buf[:])
		if r < limit {
			return min + int32(r%rangeSize)
		}
	}
}
