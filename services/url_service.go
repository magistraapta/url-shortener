package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"url-shortener/models"
	"url-shortener/repositories"
	"url-shortener/util"

	"github.com/google/uuid"
)

// ErrInvalidURL is returned by Shorten when the submitted URL is empty, too
// long, unparseable, not http(s), or points back at this service.
var ErrInvalidURL = errors.New("invalid url")

const (
	defaultCodeLen = 7
	maxURLLen      = 2048
	maxAttempts    = 3
)

type IUrlService interface {
	// Shorten validates longURL, stores a mapping for it, and returns the
	// full short URL. A URL that was shortened before reuses its code.
	Shorten(ctx context.Context, longURL string) (models.ShortUrlResponse, error)

	// Resolve returns the original URL for a short code, or
	// repositories.ErrURLNotFound if the code is unknown.
	Resolve(ctx context.Context, shortCode string) (string, error)
}

type UrlService struct {
	repo     repositories.IUrlRepository
	baseURL  string // e.g. "https://sho.rt", no trailing slash
	baseHost string // host of baseURL, for the redirect-loop guard
	codeLen  int
}

func NewUrlService(repo repositories.IUrlRepository, baseURL string) IUrlService {
	trimmed := strings.TrimRight(baseURL, "/")
	host := ""
	if u, err := url.Parse(trimmed); err == nil {
		host = u.Hostname()
	}
	return &UrlService{
		repo:     repo,
		baseURL:  trimmed,
		baseHost: host,
		codeLen:  defaultCodeLen,
	}
}

func (s *UrlService) Shorten(ctx context.Context, longURL string) (models.ShortUrlResponse, error) {
	if err := s.validate(longURL); err != nil {
		return models.ShortUrlResponse{}, err
	}

	hash := util.HashURL(longURL)

	for range maxAttempts {
		code, err := util.GenerateCode(s.codeLen)
		if err != nil {
			return models.ShortUrlResponse{}, fmt.Errorf("shorten: %w", err)
		}

		stored, err := s.repo.CreateOrGetShortCode(ctx, models.Url{
			ID:          uuid.New(),
			LongURL:     longURL,
			LongURLHash: hash,
			ShortURL:    code,
		})
		switch {
		case errors.Is(err, repositories.ErrShortCodeExists):
			continue // code collided with a different URL; regenerate
		case err != nil:
			return models.ShortUrlResponse{}, fmt.Errorf("shorten: %w", err)
		}

		return models.ShortUrlResponse{
			LongURL:  longURL,
			ShortURL: s.baseURL + "/" + stored,
		}, nil
	}

	return models.ShortUrlResponse{}, fmt.Errorf("shorten: no free code after %d attempts", maxAttempts)
}

func (s *UrlService) Resolve(ctx context.Context, shortCode string) (string, error) {
	if shortCode == "" {
		return "", repositories.ErrURLNotFound
	}
	return s.repo.GetLongURL(ctx, shortCode)
}

func (s *UrlService) validate(raw string) error {
	if raw == "" || len(raw) > maxURLLen {
		return ErrInvalidURL
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ErrInvalidURL
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ErrInvalidURL
	}
	if s.baseHost != "" && strings.EqualFold(u.Hostname(), s.baseHost) {
		return ErrInvalidURL // don't shorten our own short links
	}
	return nil
}
