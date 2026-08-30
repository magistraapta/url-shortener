package util

import "crypto/sha256"

// HashURL returns the SHA-256 digest of a long URL. It is stored in the
// url.long_url_hash column and carries the UNIQUE constraint used for
// deduplication, since the long_url TEXT column itself is too large to index.
func HashURL(longURL string) []byte {
	sum := sha256.Sum256([]byte(longURL))
	return sum[:]
}
