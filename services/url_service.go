package services

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"url-shortener/repositories"
	"url-shortener/util"
)

// ErrInvalidURL is returned by Shorten when the input is not a usable http(s) URL.
var ErrInvalidURL = errors.New("invalid url")

// normalizeURL trims raw, assumes https:// when no scheme is given, and checks
// that the result is an absolute http(s) URL with a host. Without a scheme a
// redirect would be treated by clients as a relative path.
func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidURL
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return "", ErrInvalidURL
	}
	return u.String(), nil
}

type UrlServiceInterface interface {
	Shorten(ctx context.Context, longURL string) (string, error)
	Resolve(ctx context.Context, shortCode string) (string, error)
}

type UrlService struct {
	repo repositories.UrlRepositoryInterface
}

func NewUrlService(repo repositories.UrlRepositoryInterface) *UrlService {
	return &UrlService{repo: repo}
}

// Todo: Implement helper function for generating short code
func (s *UrlService) Shorten(ctx context.Context, longURL string) (string, error) {
	// Implement the logic to generate a short code and save it using the repository
	// For example, you can use a hash function or a random string generator
	longURL, err := normalizeURL(longURL)
	if err != nil {
		return "", err
	}
	shortCode := util.ShortCodeFromURL(longURL) // Replace with actual code generation logic
	err = s.repo.SaveUrl(ctx, longURL, shortCode)
	if err != nil {
		return "", err
	}
	return shortCode, nil
}

func (s *UrlService) Resolve(ctx context.Context, shortCode string) (string, error) {
	longURL, err := s.repo.GetLongURL(ctx, shortCode)
	if err != nil {
		return "", err
	}
	return longURL, nil
}
