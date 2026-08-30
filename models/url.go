package models

import (
	"time"

	"github.com/google/uuid"
)

type Url struct {
	ID          uuid.UUID `json:"id"`
	LongURL     string    `json:"long_url"`
	LongURLHash []byte    `json:"-"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type ShortUrlRequest struct {
	LongURL string `json:"long_url"`
}

type ShortUrlResponse struct {
	LongURL  string `json:"long_url"`
	ShortURL string `json:"short_url"`
}
