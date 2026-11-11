package services

import (
	"context"
	"url-shortener/repositories"
)

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
	shortCode := "generatedShortCode" // Replace with actual code generation logic
	err := s.repo.SaveUrl(ctx, longURL, shortCode)
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
