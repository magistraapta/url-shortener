package controller

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	models "url-shortener/models"
	"url-shortener/services"
)

var (
	ErrURLNotFound = errors.New("short link not found")
)

type UrlController struct {
	svc services.UrlServiceInterface
}

func NewUrlController(svc services.UrlServiceInterface) *UrlController {
	return &UrlController{svc: svc}
}

// RegisterRoutes wires the controller's handlers onto mux.
func (c *UrlController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/urls", c.Shorten)
	mux.HandleFunc("GET /{code}", c.Redirect)
}

// Shorten handles POST /api/urls with body {"long_url": "..."} and responds
// 201 with {"long_url": "...", "short_url": "..."}.
func (c *UrlController) Shorten(w http.ResponseWriter, r *http.Request) {
	var req models.ShortUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "body must be JSON with a long_url field")
		return
	}

	res, err := c.svc.Shorten(r.Context(), req.LongURL)
	switch {
	case errors.Is(err, services.ErrInvalidURL):
		writeError(w, http.StatusBadRequest, "invalid_url", "invalid url")
		return
	case err != nil:
		slog.Error("shorten failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	writeSuccess(w, http.StatusCreated, res)
}

// Redirect handles GET /{code} and responds 302 to the original URL.
func (c *UrlController) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	longURL, err := c.svc.Resolve(r.Context(), code)
	switch {
	case errors.Is(err, ErrURLNotFound):
		writeError(w, http.StatusNotFound, "not_found", "short link not found")
		return
	case err != nil:
		slog.Error("resolve failed", "err", err, "code", code)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func writeSuccess[T any](w http.ResponseWriter, status int, data T) {
	writeJSON(w, status, models.Success(data))
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, models.Failure(code, message))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write response failed", "err", err)
	}
}
