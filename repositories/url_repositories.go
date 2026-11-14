package repositories

import (
	"context"
	models "url-shortener/models"
	"url-shortener/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		ID:          uuid.New(),
		LongURL:     longURL,
		LongURLHash: util.HashURL(longURL),
		ShortURL:    shortCode,
	}
	// The same long URL is already stored: nothing to do. The short code is
	// derived deterministically from the URL, so the existing row has the same
	// code. A short_url collision between *different* URLs is not covered by
	// this target and still surfaces as an error.
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "long_url_hash"}},
			DoNothing: true,
		}).
		Create(&url).Error
}

func (r *UrlRepository) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	var url models.Url
	result := r.db.WithContext(ctx).Where("short_url = ?", shortCode).First(&url)
	if result.Error != nil {
		return "", result.Error
	}
	return url.LongURL, nil
}
