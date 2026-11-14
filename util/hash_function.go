package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
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

// ShortCodeLength is the fixed length of codes returned by ShortCodeFromURL.
const ShortCodeLength = 7

// shortCodeSpace is len(codeAlphabet)^ShortCodeLength (62^7), the number of
// distinct codes. Reducing a uint64 modulo it leaves a bias of roughly 1 in 5
// million, which is negligible for a short-code.
const shortCodeSpace = 62 * 62 * 62 * 62 * 62 * 62 * 62

// ShortCodeFromURL deterministically hashes longURL into a 7-character
// alphanumeric code (0-9, A-Z, a-z). The same URL always yields the same code.
// Different URLs can collide (62^7 ≈ 3.5e12 codes), so callers should rely on
// the unique constraint on short_url and handle a conflict.
func ShortCodeFromURL(longURL string) string {
	sum := HashURL(longURL)
	n := binary.BigEndian.Uint64(sum[:8]) % shortCodeSpace

	code := make([]byte, ShortCodeLength)
	for i := ShortCodeLength - 1; i >= 0; i-- {
		code[i] = codeAlphabet[n%62]
		n /= 62
	}
	return string(code)
}
