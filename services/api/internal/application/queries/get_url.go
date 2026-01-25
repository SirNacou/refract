package queries

import (
	"context"
	"time"

	"github.com/SirNacou/refract/services/api/internal/application/service"
	"github.com/SirNacou/refract/services/api/internal/domain"
)

type GetURLByShortCodeQuery struct {
	ShortCode string
}

type GetURLByShortCodeResult struct {
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	Title       string     `json:"title"`
	Notes       string     `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	OwnerID     string     `json:"owner_id,omitempty"`
}

type GetURLByShortCodeHandler struct {
	store domain.Store
	cache service.Cache
}

func NewGetURLByShortCodeHandler(
	store domain.Store,
	cache service.Cache,
) *GetURLByShortCodeHandler {
	return &GetURLByShortCodeHandler{
		store,
		cache,
	}
}

func (h *GetURLByShortCodeHandler) Handle(
	ctx context.Context,
	query GetURLByShortCodeQuery,
) (*GetURLByShortCodeResult, error) {

	url, err := h.store.URLs().GetByCustomAlias(ctx, query.ShortCode)
	if err != nil {
		return nil, err
	}
	return &GetURLByShortCodeResult{
		ShortCode:   url.ShortCode.String(),
		OriginalURL: url.DestinationURL,
		CreatedAt:   url.CreatedAt,
		ExpiresAt:   url.ExpiresAt,
		OwnerID:     url.CreatorUserID,
	}, nil
}
