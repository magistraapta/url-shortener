package services

import (
	"errors"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"adds https when scheme missing", "youtube.com", "https://youtube.com", false},
		{"keeps path and query", "x.com/a/b?c=1", "https://x.com/a/b?c=1", false},
		{"keeps https", "https://example.com/p", "https://example.com/p", false},
		{"keeps http", "http://example.com", "http://example.com", false},
		{"trims whitespace", "  example.com  ", "https://example.com", false},
		{"empty", "", "", true},
		{"blank", "   ", "", true},
		{"unsupported scheme", "ftp://example.com", "", true},
		{"javascript scheme", "javascript://alert(1)", "", true},
		{"no host", "https://", "", true},
		{"spaces inside", "exa mple.com", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeURL(tt.in)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidURL) {
					t.Fatalf("normalizeURL(%q) err = %v, want ErrInvalidURL", tt.in, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeURL(%q) unexpected err: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
