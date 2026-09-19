package util

import (
	"crypto/rand"
	"fmt"
)

// base62Alphabet is ordered digits, uppercase, lowercase. All 62 characters are
// URL-safe with no escaping.
const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// base62Limit is the largest multiple of 62 that fits in a byte (248). Random
// bytes at or above it are discarded so that b % 62 stays uniform (no modulo
// bias toward the first 256 % 62 = 8 characters).
const base62Limit = 256 - (256 % 62)

// GenerateCode returns a cryptographically random base62 string of length n,
// suitable as a short-URL code: URL-safe and impractical to enumerate.
//
// Collisions are still possible in principle, so the caller must persist the
// code under a UNIQUE constraint and retry on violation.
func GenerateCode(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("generate code: length must be positive, got %d", n)
	}

	out := make([]byte, n)
	buf := make([]byte, n)
	filled := 0
	for filled < n {
		if _, err := rand.Read(buf); err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}
		for _, b := range buf {
			if int(b) >= base62Limit {
				continue // reject to avoid modulo bias
			}
			out[filled] = base62Alphabet[b%62]
			filled++
			if filled == n {
				break
			}
		}
	}
	return string(out), nil
}
