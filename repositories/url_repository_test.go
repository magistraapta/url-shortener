package repositories

import (
	"context"
	"errors"
	"testing"

	"url-shortener/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeRow / fakeDB implement the minimal pgx surface this repository uses.
type fakeRow struct {
	scan func(dest ...any) error
}

func (r fakeRow) Scan(dest ...any) error { return r.scan(dest...) }

type fakeDB struct {
	row fakeRow
}

func (d fakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (d fakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (d fakeDB) QueryRow(context.Context, string, ...any) pgx.Row        { return d.row }

func newRepo(scan func(dest ...any) error) *URLRepository {
	return &URLRepository{db: fakeDB{row: fakeRow{scan: scan}}}
}

func setString(dest []any, v string) {
	*dest[0].(*string) = v
}

func TestGetLongURL_OK(t *testing.T) {
	repo := newRepo(func(dest ...any) error {
		setString(dest, "https://example.com/dest")
		return nil
	})

	got, err := repo.GetLongURL(context.Background(), "abc1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/dest" {
		t.Errorf("got %q", got)
	}
}

func TestGetLongURL_NoRowsMapsToNotFound(t *testing.T) {
	repo := newRepo(func(...any) error { return pgx.ErrNoRows })

	_, err := repo.GetLongURL(context.Background(), "missing")
	if !errors.Is(err, ErrURLNotFound) {
		t.Fatalf("error = %v, want ErrURLNotFound", err)
	}
}

func TestGetLongURL_OtherErrorIsWrapped(t *testing.T) {
	boom := errors.New("connection reset")
	repo := newRepo(func(...any) error { return boom })

	_, err := repo.GetLongURL(context.Background(), "abc")
	if errors.Is(err, ErrURLNotFound) {
		t.Fatalf("error = %v, should not be ErrURLNotFound", err)
	}
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want it to wrap %v", err, boom)
	}
}

func TestCreateOrGetShortCode_OK(t *testing.T) {
	repo := newRepo(func(dest ...any) error {
		setString(dest, "abc1234")
		return nil
	})

	got, err := repo.CreateOrGetShortCode(context.Background(), models.Url{ShortURL: "abc1234"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc1234" {
		t.Errorf("got %q, want abc1234", got)
	}
}

func TestCreateOrGetShortCode_ShortCodeCollision(t *testing.T) {
	repo := newRepo(func(...any) error {
		return &pgconn.PgError{Code: "23505", ConstraintName: shortURLUniqueConstraint}
	})

	_, err := repo.CreateOrGetShortCode(context.Background(), models.Url{})
	if !errors.Is(err, ErrShortCodeExists) {
		t.Fatalf("error = %v, want ErrShortCodeExists", err)
	}
}

func TestCreateOrGetShortCode_OtherUniqueViolationIsNotCollision(t *testing.T) {
	// A 23505 on a different constraint must not be mistaken for a code clash.
	repo := newRepo(func(...any) error {
		return &pgconn.PgError{Code: "23505", ConstraintName: "url_long_url_hash_key"}
	})

	_, err := repo.CreateOrGetShortCode(context.Background(), models.Url{})
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrShortCodeExists) {
		t.Fatalf("error = %v, should not be ErrShortCodeExists", err)
	}
}

func TestCreateOrGetShortCode_GenericErrorIsWrapped(t *testing.T) {
	boom := errors.New("deadline exceeded")
	repo := newRepo(func(...any) error { return boom })

	_, err := repo.CreateOrGetShortCode(context.Background(), models.Url{})
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want it to wrap %v", err, boom)
	}
}
