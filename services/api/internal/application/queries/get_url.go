package queries

import (
	"context"
	"time"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
)

// GetURLQuery represents the query to get URL metadata
type GetURLQuery struct {
	ShortCode string
}

// GetURLResult contains the URL metadata
type GetURLResult struct {
	ShortCode          string
	OriginalURL        string
	Domain             string
	CreatedAt          time.Time
	ExpiresAt          time.Time
	HasFixedExpiration bool
	ClickCount         int64
	IsActive           bool
}

// GetURLHandler handles the GetURL query
type GetURLHandler struct {
	repo url.Repository
}

// NewGetURLHandler creates a new GetURLHandler
func NewGetURLHandler(repo url.Repository) *GetURLHandler {
	return &GetURLHandler{
		repo: repo,
	}
}

// Handle executes the GetURL query
func (h *GetURLHandler) Handle(ctx context.Context, query GetURLQuery) (*GetURLResult, error) {
	// Validate short code
	shortCode, err := url.NewShortCode(query.ShortCode)
	if err != nil {
		return nil, err
	}

	// Find URL by short code
	urlEntity, err := h.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// Check if expired
	if urlEntity.IsExpired() {
		return nil, url.NewNotFoundError("URL_EXPIRED", "This URL has expired")
	}

	// Check if active
	if !urlEntity.IsActive() {
		return nil, url.NewNotFoundError("URL_INACTIVE", "This URL is no longer active")
	}

	return &GetURLResult{
		ShortCode:          urlEntity.ShortCode().String(),
		OriginalURL:        urlEntity.OriginalURL().String(),
		Domain:             urlEntity.Domain().String(),
		CreatedAt:          urlEntity.CreatedAt(),
		ExpiresAt:          urlEntity.ExpiresAt(),
		HasFixedExpiration: urlEntity.HasFixedExpiration(),
		ClickCount:         urlEntity.ClickCount(),
		IsActive:           urlEntity.IsActive(),
	}, nil
}
