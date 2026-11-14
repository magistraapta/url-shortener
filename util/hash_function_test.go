package util

import (
	"regexp"
	"testing"
)

func TestShortCodeFromURL(t *testing.T) {
	valid := regexp.MustCompile(`^[0-9A-Za-z]{7}$`)

	urls := []string{
		"https://example.com",
		"https://example.com/a/very/long/path?with=query&and=more#fragment",
		"",
		"ünïcödé://例え.jp/パス",
	}
	for _, u := range urls {
		got := ShortCodeFromURL(u)
		if !valid.MatchString(got) {
			t.Errorf("ShortCodeFromURL(%q) = %q, want 7 alphanumeric chars", u, got)
		}
		if again := ShortCodeFromURL(u); again != got {
			t.Errorf("ShortCodeFromURL(%q) not deterministic: %q vs %q", u, got, again)
		}
	}

	if ShortCodeFromURL("https://example.com/a") == ShortCodeFromURL("https://example.com/b") {
		t.Error("different URLs produced the same code")
	}
}
