package util

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateCode_Length(t *testing.T) {
	for _, n := range []int{1, 7, 22, 128} {
		got, err := GenerateCode(n)
		if err != nil {
			t.Fatalf("GenerateCode(%d): %v", n, err)
		}
		if len(got) != n {
			t.Errorf("GenerateCode(%d) length = %d, want %d", n, len(got), n)
		}
	}
}

func TestGenerateCode_OnlyBase62Chars(t *testing.T) {
	got, err := GenerateCode(4096)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range got {
		if !strings.ContainsRune(base62Alphabet, r) {
			t.Fatalf("character %q is not in the base62 alphabet", r)
		}
	}
}

func TestGenerateCode_NonPositiveLength(t *testing.T) {
	for _, n := range []int{0, -1, -100} {
		if _, err := GenerateCode(n); err == nil {
			t.Errorf("GenerateCode(%d) error = nil, want non-nil", n)
		}
	}
}

func TestGenerateCode_NoObviousCollisions(t *testing.T) {
	const draws = 5000
	seen := make(map[string]struct{}, draws)
	for range draws {
		c, err := GenerateCode(7)
		if err != nil {
			t.Fatal(err)
		}
		if _, dup := seen[c]; dup {
			t.Fatalf("duplicate code %q within %d draws", c, draws)
		}
		seen[c] = struct{}{}
	}
}

func TestHashURL(t *testing.T) {
	a := HashURL("https://example.com/path")
	b := HashURL("https://example.com/path")
	c := HashURL("https://example.com/other")

	if len(a) != 32 {
		t.Errorf("hash length = %d, want 32 (sha-256)", len(a))
	}
	if !bytes.Equal(a, b) {
		t.Error("same input produced different hashes")
	}
	if bytes.Equal(a, c) {
		t.Error("different inputs produced the same hash")
	}
}
