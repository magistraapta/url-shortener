package util

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
)

func HashURL(longURL string) []byte {
	sum := sha256.Sum256([]byte(longURL))
	return sum[:]
}

const codeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func GenerateCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("generate code: length must be positive")
	}

	// Largest multiple of len(alphabet) that fits in a byte; bytes at or above
	// it are discarded so every character is equally likely (no modulo bias).
	const maxUnbiased = 256 - 256%len(codeAlphabet)

	code := make([]byte, 0, length)
	buf := make([]byte, length*2)
	for len(code) < length {
		if _, err := rand.Read(buf); err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}
		for _, b := range buf {
			if int(b) >= maxUnbiased {
				continue
			}
			code = append(code, codeAlphabet[int(b)%len(codeAlphabet)])
			if len(code) == length {
				break
			}
		}
	}
	return string(code), nil
}
