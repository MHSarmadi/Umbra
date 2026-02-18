package math_tools

import (
	"crypto/rand"
)

func RandomDecimalString(n int) string {
	if n <= 0 {
		return ""
	}

	buf := make([]byte, n)
	for i := range n {
		// pick a number 0-9
		var b [1]byte
		for {
			if _, err := rand.Read(b[:]); err != nil {
				panic(err)
			}
			if b[0] < 250 { // 250 = largest multiple of 10 < 256
				buf[i] = '0' + (b[0] % 10)
				break
			}
		}
	}

	return string(buf)
}
