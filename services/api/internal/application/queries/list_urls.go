package queries

import (
	"context"

	"github.com/SirNacou/refract/services/api/internal/application/service"
	"github.com/SirNacou/refract/services/api/internal/domain"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/middleware"
)

type ListURLsQuery struct {
	// Potentially add filters, pagination, etc.
}

type ListURLsResult struct {
	URLs []GetURLByShortCodeResult `json:"urls"`
}

type ListURLsHandler struct {
	store domain.Store
	cache service.Cache
}

func NewListURLsHandler(
	store domain.Store,
	cache service.Cache,
) *ListURLsHandler {
	return &ListURLsHandler{
		store,
		cache,
	}
}

func (h *ListURLsHandler) Handle(
	ctx context.Context,
	query ListURLsQuery,
) (*ListURLsResult, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, middleware.ErrUserIDNotFound
	}

	// Implementation to list URLs from the store
	urls, err := h.store.URLs().GetByCreatorID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	var results []GetURLByShortCodeResult
	for _, url := range urls {
		results = append(results, GetURLByShortCodeResult{
			ShortCode:   url.ShortCode.String(),
			OriginalURL: url.DestinationURL,
			Title:       url.Title,
			Notes:       url.Notes,
			CreatedAt:   url.CreatedAt,
			ExpiresAt:   url.ExpiresAt,
			OwnerID:     url.CreatorUserID,
		})
	}
	return &ListURLsResult{
		URLs: results,
	}, nil
}
