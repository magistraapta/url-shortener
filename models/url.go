package model

import "github.com/google/uuid"

type Url struct {
	ID          uuid.UUID `json:"id"`
	LongURL     string    `json:"long_url"`
	LongURLHash []byte    `json:"-"` // dedup key (SHA-256 of LongURL)
	ShortURL    string    `json:"short_url"`
}

type ShortUrlRequest struct {
	LongURL string `json:"long_url"`
}

type ShortUrlResponse struct {
	LongURL  string `json:"long_url"`
	ShortURL string `json:"short_url"`
}

func (Url) TableName() string { return "url" }
