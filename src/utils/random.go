package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func RandomAlphaNumeric(n int) (string, error) {
	if n < 0 {
		return "", fmt.Errorf("n must be non-negative")
	}
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		// 0 <= x < len(letters)
		x, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[x.Int64()]
	}
	return string(b), nil
}
