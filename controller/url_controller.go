package controller

import (
	"errors"
	"log/slog"
	"net/http"

	"url-shortener/models"
	"url-shortener/repositories"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
)

type UrlController struct {
	svc services.IUrlService
}

func NewUrlController(svc services.IUrlService) *UrlController {
	return &UrlController{svc: svc}
}

// RegisterRoutes wires the controller's handlers onto r.
func (c *UrlController) RegisterRoutes(r gin.IRouter) {
	r.POST("/api/urls", c.Shorten)
	r.GET("/:code", c.Redirect)
}

// Shorten handles POST /api/urls with body {"long_url": "..."} and responds
// 201 with {"long_url": "...", "short_url": "..."}.
func (c *UrlController) Shorten(ctx *gin.Context) {
	var req models.ShortUrlRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "body must be JSON with a long_url field"})
		return
	}

	res, err := c.svc.Shorten(ctx.Request.Context(), req.LongURL)
	switch {
	case errors.Is(err, services.ErrInvalidURL):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	case err != nil:
		slog.Error("shorten failed", "err", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}

// Redirect handles GET /:code and 302-redirects to the original URL.
func (c *UrlController) Redirect(ctx *gin.Context) {
	code := ctx.Param("code")

	longURL, err := c.svc.Resolve(ctx.Request.Context(), code)
	switch {
	case errors.Is(err, repositories.ErrURLNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "short link not found"})
		return
	case err != nil:
		slog.Error("resolve failed", "err", err, "code", code)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	ctx.Redirect(http.StatusFound, longURL)
}
