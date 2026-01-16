package commands

import (
	"time"

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
	generator *idgen.SnowflakeGenerator
}

func (h *CreateURLHandler) Handle(cmd CreateURLCommand) (*CreateURLResult, error) {
	

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

	return &CreateURLResult{
		ShortCode: r.CustomAlias.String(),
	}, nil
}
