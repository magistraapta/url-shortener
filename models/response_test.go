package model

import (
	"encoding/json"
	"testing"
)

func TestResponseJSON(t *testing.T) {
	t.Run("success carries data and no error", func(t *testing.T) {
		b, err := json.Marshal(Success(ShortUrlResponse{LongURL: "https://x.com", ShortURL: "abc"}))
		if err != nil {
			t.Fatal(err)
		}
		want := `{"success":true,"data":{"long_url":"https://x.com","short_url":"abc"}}`
		if string(b) != want {
			t.Fatalf("got %s, want %s", b, want)
		}
	})

	t.Run("success works with a bare string payload", func(t *testing.T) {
		b, _ := json.Marshal(Success("abc"))
		want := `{"success":true,"data":"abc"}`
		if string(b) != want {
			t.Fatalf("got %s, want %s", b, want)
		}
	})

	t.Run("failure carries error and no data", func(t *testing.T) {
		b, err := json.Marshal(Failure("invalid_url", "invalid url"))
		if err != nil {
			t.Fatal(err)
		}
		want := `{"success":false,"error":{"code":"invalid_url","message":"invalid url"}}`
		if string(b) != want {
			t.Fatalf("got %s, want %s", b, want)
		}
	})
}
