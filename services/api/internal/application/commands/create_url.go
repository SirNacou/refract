package commands

import (
	"context"
	"time"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/safebrowsing"
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
	generator *idgen.SnowflakeGenerator
	sb        *safebrowsing.SafeBrowsing
	urlRepo   url.URLRepository
}

func NewCreateURLHandler(generator *idgen.SnowflakeGenerator, sb *safebrowsing.SafeBrowsing, urlRepo url.URLRepository) *CreateURLHandler {
	return &CreateURLHandler{
		generator: generator,
		sb:        sb,
		urlRepo:   urlRepo,
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

	var customAlias *url.ShortCode = nil
	if cmd.CustomAlias != nil {
		shortCode, err := url.NewCustomShortCode(*cmd.CustomAlias)
		if err != nil {
			return nil, err
		}
		customAlias = shortCode
	}

	r, err := url.NewURL(url.CreateURLRequest{
		ID:             id,
		CustomAlias:    customAlias,
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
	if err := h.urlRepo.Create(r); err != nil {
		return nil, err
	}

	return &CreateURLResult{
		ShortCode: r.CustomAlias.String(),
	}, nil
}
