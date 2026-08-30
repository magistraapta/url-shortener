package repositories

import (
	"context"
	"errors"
	"fmt"
	"url-shortener/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrURLNotFound is returned when no row matches the lookup. Callers in the
// service/handler layer can check for this without importing pgx.
var ErrURLNotFound = errors.New("url not found")

// ErrShortCodeExists is returned when the generated short code collides with a
// code already stored for a different URL. The service is expected to generate a
// new code and retry.
var ErrShortCodeExists = errors.New("short code already exists")

// shortURLUniqueConstraint is the name of the UNIQUE constraint on url.short_url
// (see the create_url_table migration). A 23505 error naming it means the code
// collided rather than the long_url_hash.
const shortURLUniqueConstraint = "url_short_url_key"

// DB is the subset of pgx used by this repository. Both *pgx.Conn and
// *pgxpool.Pool satisfy it, and it can be faked in unit tests.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type IUrlRepository interface {
	GetLongURL(ctx context.Context, shortCode string) (string, error)
	CreateOrGetShortCode(ctx context.Context, u models.Url) (string, error)
}

type UrlRepository struct {
	db DB
}

func NewUrlRepository(db DB) IUrlRepository {
	return &UrlRepository{db: db}
}

// GetLongURL resolves a short code to its original URL for redirects.
// It returns ErrURLNotFound if there is no such row.
func (r *UrlRepository) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	const query = `SELECT long_url FROM url WHERE short_url = $1 LIMIT 1`

	var longURL string
	err := r.db.QueryRow(ctx, query, shortCode).Scan(&longURL)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "", ErrURLNotFound
	case err != nil:
		return "", fmt.Errorf("get long url: %w", err)
	}
	return longURL, nil
}

// CreateOrGetShortCode inserts u as a new mapping. If u.LongURLHash already
// exists, it returns the short code stored on that existing row instead. Either
// way the returned string is the code the caller should serve.
func (r *UrlRepository) CreateOrGetShortCode(ctx context.Context, u models.Url) (string, error) {
	const query = `
		INSERT INTO url (id, long_url, long_url_hash, short_url, created_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (long_url_hash) DO UPDATE SET long_url = EXCLUDED.long_url
		RETURNING short_url`

	var shortCode string
	err := r.db.QueryRow(ctx, query,
		u.ID, u.LongURL, u.LongURLHash, u.ShortURL,
	).Scan(&shortCode)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == shortURLUniqueConstraint {
			return "", ErrShortCodeExists
		}
		return "", fmt.Errorf("create or get short code: %w", err)
	}
	return shortCode, nil
}
