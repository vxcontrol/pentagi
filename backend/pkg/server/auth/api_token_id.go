package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	Base62Chars   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	TokenIDLength = 10
)

// GenerateTokenID generates a random base62 string of specified length
func GenerateTokenID() (string, error) {
	b := make([]byte, TokenIDLength)
	maxIdx := big.NewInt(int64(len(Base62Chars)))

	for i := range b {
		idx, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			return "", fmt.Errorf("error generating token ID: %w", err)
		}
		b[i] = Base62Chars[idx.Int64()]
	}

	return string(b), nil
}
