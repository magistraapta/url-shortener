package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"url-shortener/models"
	"url-shortener/repositories"
)

// fakeRepo is a hand-written stub for repositories.IUrlRepository.
type fakeRepo struct {
	createFn    func(ctx context.Context, u models.Url) (string, error)
	getFn       func(ctx context.Context, code string) (string, error)
	createCalls int
	getCalls    int
}

func (f *fakeRepo) CreateOrGetShortCode(ctx context.Context, u models.Url) (string, error) {
	f.createCalls++
	return f.createFn(ctx, u)
}

func (f *fakeRepo) GetLongURL(ctx context.Context, code string) (string, error) {
	f.getCalls++
	return f.getFn(ctx, code)
}

func failCreate(t *testing.T) func(context.Context, models.Url) (string, error) {
	return func(context.Context, models.Url) (string, error) {
		t.Helper()
		t.Fatal("CreateOrGetShortCode should not have been called")
		return "", nil
	}
}

func failGet(t *testing.T) func(context.Context, string) (string, error) {
	return func(context.Context, string) (string, error) {
		t.Helper()
		t.Fatal("GetLongURL should not have been called")
		return "", nil
	}
}

func TestShorten_Valid(t *testing.T) {
	repo := &fakeRepo{
		getFn: failGet(t),
		createFn: func(_ context.Context, u models.Url) (string, error) {
			if u.LongURL != "https://example.com/page" {
				t.Errorf("repo got LongURL %q", u.LongURL)
			}
			if len(u.LongURLHash) != 32 {
				t.Errorf("repo got hash of length %d, want 32", len(u.LongURLHash))
			}
			if len(u.ShortURL) != defaultCodeLen {
				t.Errorf("repo got code %q of length %d, want %d", u.ShortURL, len(u.ShortURL), defaultCodeLen)
			}
			return u.ShortURL, nil
		},
	}
	svc := NewUrlService(repo, "https://sho.rt/")

	res, err := svc.Shorten(context.Background(), "https://example.com/page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.LongURL != "https://example.com/page" {
		t.Errorf("LongURL = %q", res.LongURL)
	}
	if !strings.HasPrefix(res.ShortURL, "https://sho.rt/") {
		t.Errorf("ShortURL = %q, want https://sho.rt/ prefix", res.ShortURL)
	}
	if code := strings.TrimPrefix(res.ShortURL, "https://sho.rt/"); len(code) != defaultCodeLen {
		t.Errorf("code %q length = %d, want %d", code, len(code), defaultCodeLen)
	}
}

func TestShorten_ReusesExistingCode(t *testing.T) {
	repo := &fakeRepo{
		getFn: failGet(t),
		createFn: func(context.Context, models.Url) (string, error) {
			return "EXISTING", nil // dedup: repo returns a pre-existing code
		},
	}
	svc := NewUrlService(repo, "https://sho.rt")

	res, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ShortURL != "https://sho.rt/EXISTING" {
		t.Errorf("ShortURL = %q, want https://sho.rt/EXISTING", res.ShortURL)
	}
}

func TestShorten_InvalidURL(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", maxURLLen)
	cases := map[string]string{
		"empty":          "",
		"no scheme":      "example.com",
		"path only":      "notaurl",
		"unsupported":    "ftp://example.com",
		"scheme no host": "http://",
		"has spaces":     "http://exa mple.com",
		"too long":       long,
		"own domain":     "https://sho.rt/abc1234",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepo{createFn: failCreate(t), getFn: failGet(t)}
			svc := NewUrlService(repo, "https://sho.rt")

			_, err := svc.Shorten(context.Background(), in)
			if !errors.Is(err, ErrInvalidURL) {
				t.Fatalf("error = %v, want ErrInvalidURL", err)
			}
			if repo.createCalls != 0 {
				t.Errorf("repo called %d times, want 0", repo.createCalls)
			}
		})
	}
}

func TestShorten_RetriesOnCodeCollision(t *testing.T) {
	repo := &fakeRepo{getFn: failGet(t)}
	repo.createFn = func(_ context.Context, u models.Url) (string, error) {
		if repo.createCalls < 3 {
			return "", repositories.ErrShortCodeExists
		}
		return u.ShortURL, nil
	}
	svc := NewUrlService(repo, "https://sho.rt")

	res, err := svc.Shorten(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalls != 3 {
		t.Errorf("createCalls = %d, want 3", repo.createCalls)
	}
	if !strings.HasPrefix(res.ShortURL, "https://sho.rt/") {
		t.Errorf("ShortURL = %q", res.ShortURL)
	}
}

func TestShorten_CollisionRetriesExhausted(t *testing.T) {
	repo := &fakeRepo{
		getFn:    failGet(t),
		createFn: func(context.Context, models.Url) (string, error) { return "", repositories.ErrShortCodeExists },
	}
	svc := NewUrlService(repo, "https://sho.rt")

	_, err := svc.Shorten(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("expected an error after retries are exhausted")
	}
	if errors.Is(err, ErrInvalidURL) {
		t.Errorf("error = %v, should not be ErrInvalidURL", err)
	}
	if repo.createCalls != maxAttempts {
		t.Errorf("createCalls = %d, want %d", repo.createCalls, maxAttempts)
	}
}

func TestShorten_RepoError(t *testing.T) {
	sentinel := errors.New("db exploded")
	repo := &fakeRepo{
		getFn:    failGet(t),
		createFn: func(context.Context, models.Url) (string, error) { return "", sentinel },
	}
	svc := NewUrlService(repo, "https://sho.rt")

	_, err := svc.Shorten(context.Background(), "https://example.com")
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want it to wrap %v", err, sentinel)
	}
}

func TestResolve_OK(t *testing.T) {
	repo := &fakeRepo{
		createFn: failCreate(t),
		getFn: func(_ context.Context, code string) (string, error) {
			if code != "abc1234" {
				t.Errorf("repo got code %q", code)
			}
			return "https://example.com/dest", nil
		},
	}
	svc := NewUrlService(repo, "https://sho.rt")

	got, err := svc.Resolve(context.Background(), "abc1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/dest" {
		t.Errorf("got %q", got)
	}
}

func TestResolve_NotFound(t *testing.T) {
	repo := &fakeRepo{
		createFn: failCreate(t),
		getFn:    func(context.Context, string) (string, error) { return "", repositories.ErrURLNotFound },
	}
	svc := NewUrlService(repo, "https://sho.rt")

	_, err := svc.Resolve(context.Background(), "missing")
	if !errors.Is(err, repositories.ErrURLNotFound) {
		t.Fatalf("error = %v, want ErrURLNotFound", err)
	}
}

func TestResolve_EmptyCode(t *testing.T) {
	repo := &fakeRepo{createFn: failCreate(t), getFn: failGet(t)}
	svc := NewUrlService(repo, "https://sho.rt")

	_, err := svc.Resolve(context.Background(), "")
	if !errors.Is(err, repositories.ErrURLNotFound) {
		t.Fatalf("error = %v, want ErrURLNotFound", err)
	}
	if repo.getCalls != 0 {
		t.Errorf("repo called %d times, want 0", repo.getCalls)
	}
}
