package commands

import (
	"context"
	"log/slog"
	"time"

	"github.com/SirNacou/refract/services/api/internal/application/cachekeys"
	"github.com/SirNacou/refract/services/api/internal/application/service"
	"github.com/SirNacou/refract/services/api/internal/domain"
	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
)

type CreateURLCommand struct {
	CustomAlias    *string
	DestinationURL string
	Title          string
	Notes          string
	ExpiresAt      *time.Time
	CreatorUserID  string
}

type CreateURLResult struct {
	ShortCode string
}

type CreateURLHandler struct {
	generator idgen.IDGenerator
	sb        service.SafeBrowsing
	store     domain.Store
	cache     service.Cache
}

func NewCreateURLHandler(generator idgen.IDGenerator, sb service.SafeBrowsing, store domain.Store,
	cache service.Cache) *CreateURLHandler {
	return &CreateURLHandler{
		generator,
		sb,
		store,
		cache,
	}
}

func (h *CreateURLHandler) Handle(ctx context.Context, cmd CreateURLCommand) (*CreateURLResult, error) {

	ok, err := h.sb.CheckURLv5Proto(ctx, cmd.DestinationURL)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, url.ErrMaliciousURL
	}

	id, err := h.generator.NextID()
	if err != nil {
		return nil, err
	}

	var customAlias *url.ShortCode
	if cmd.CustomAlias != nil {
		shortCode, err := url.NewCustomShortCode(*cmd.CustomAlias)
		if err != nil {
			return nil, err
		}
		customAlias = shortCode
	}

	r, err := url.NewURL(url.CreateURLRequest{
		ID:             id,
		ShortCode:      customAlias,
		DestinationURL: cmd.DestinationURL,
		Title:          cmd.Title,
		Notes:          cmd.Notes,
		ExpiresAt:      cmd.ExpiresAt,
		CreatorUserID:  cmd.CreatorUserID,
	})

	if err != nil {
		return nil, err
	}

	// Persist URL to database
	if err := h.store.URLs().Create(ctx, r); err != nil {
		return nil, err
	}

	err = h.cache.Client().Do(ctx,
		h.cache.Client().B().Set().
			Key(cachekeys.RedirectCacheKey(r.ShortCode)).
			Value(r.DestinationURL).Build()).
		Error()

	if err != nil {
		slog.WarnContext(ctx, "failed to warming cache when create short url",
			"short_code", r.ShortCode,
			"error", err)
	}

	return &CreateURLResult{
		ShortCode: r.ShortCode.String(),
	}, nil
}
