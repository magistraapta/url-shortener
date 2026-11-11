package repositories

import (
	"context"
	models "url-shortener/models"

	"gorm.io/gorm"
)

type UrlRepositoryInterface interface {
	SaveUrl(ctx context.Context, longURL string, shortCode string) error
	GetLongURL(ctx context.Context, shortCode string) (string, error)
}

type UrlRepository struct {
	db *gorm.DB
}

func NewUrlRepository(db *gorm.DB) *UrlRepository {
	return &UrlRepository{db: db}
}

func (r *UrlRepository) SaveUrl(ctx context.Context, longURL string, shortCode string) error {
	url := models.Url{
		LongURL:  longURL,
		ShortURL: shortCode,
	}
	return r.db.WithContext(ctx).Create(&url).Error
}

func (r *UrlRepository) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	var url models.Url
	result := r.db.WithContext(ctx).Where("short_url = ?", shortCode).First(&url)
	if result.Error != nil {
		return "", result.Error
	}
	return url.LongURL, nil
}
