package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"url-shortener/models"
	"url-shortener/repositories"
	"url-shortener/services"
)

type fakeSvc struct {
	shortenFn func(ctx context.Context, longURL string) (models.ShortUrlResponse, error)
	resolveFn func(ctx context.Context, code string) (string, error)
}

func (f fakeSvc) Shorten(ctx context.Context, longURL string) (models.ShortUrlResponse, error) {
	return f.shortenFn(ctx, longURL)
}

func (f fakeSvc) Resolve(ctx context.Context, code string) (string, error) {
	return f.resolveFn(ctx, code)
}

func newRouter(svc services.IUrlService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewURLController(svc).RegisterRoutes(r)
	return r
}

func do(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestShorten_Created(t *testing.T) {
	svc := fakeSvc{shortenFn: func(_ context.Context, longURL string) (models.ShortUrlResponse, error) {
		if longURL != "https://example.com" {
			t.Errorf("service got %q", longURL)
		}
		return models.ShortUrlResponse{LongURL: longURL, ShortURL: "http://localhost:8080/abc1234"}, nil
	}}
	req := httptest.NewRequest(http.MethodPost, "/api/urls",
		strings.NewReader(`{"long_url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")

	w := do(newRouter(svc), req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body)
	}
	var got models.ShortUrlResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if got.ShortURL != "http://localhost:8080/abc1234" {
		t.Errorf("short_url = %q", got.ShortURL)
	}
}

func TestShorten_BadJSON(t *testing.T) {
	svc := fakeSvc{shortenFn: func(context.Context, string) (models.ShortUrlResponse, error) {
		t.Fatal("service should not be called for invalid JSON")
		return models.ShortUrlResponse{}, nil
	}}
	req := httptest.NewRequest(http.MethodPost, "/api/urls", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")

	w := do(newRouter(svc), req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestShorten_InvalidURL(t *testing.T) {
	svc := fakeSvc{shortenFn: func(context.Context, string) (models.ShortUrlResponse, error) {
		return models.ShortUrlResponse{}, services.ErrInvalidURL
	}}
	req := httptest.NewRequest(http.MethodPost, "/api/urls",
		strings.NewReader(`{"long_url":"nonsense"}`))
	req.Header.Set("Content-Type", "application/json")

	w := do(newRouter(svc), req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestShorten_InternalError(t *testing.T) {
	svc := fakeSvc{shortenFn: func(context.Context, string) (models.ShortUrlResponse, error) {
		return models.ShortUrlResponse{}, errors.New("boom")
	}}
	req := httptest.NewRequest(http.MethodPost, "/api/urls",
		strings.NewReader(`{"long_url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")

	w := do(newRouter(svc), req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

func TestRedirect_Found(t *testing.T) {
	svc := fakeSvc{resolveFn: func(_ context.Context, code string) (string, error) {
		if code != "abc1234" {
			t.Errorf("service got code %q", code)
		}
		return "https://example.com/destination", nil
	}}
	req := httptest.NewRequest(http.MethodGet, "/abc1234", nil)

	w := do(newRouter(svc), req)

	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com/destination" {
		t.Errorf("Location = %q", loc)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	svc := fakeSvc{resolveFn: func(context.Context, string) (string, error) {
		return "", repositories.ErrURLNotFound
	}}
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)

	w := do(newRouter(svc), req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestRedirect_InternalError(t *testing.T) {
	svc := fakeSvc{resolveFn: func(context.Context, string) (string, error) {
		return "", errors.New("boom")
	}}
	req := httptest.NewRequest(http.MethodGet, "/abc1234", nil)

	w := do(newRouter(svc), req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}
