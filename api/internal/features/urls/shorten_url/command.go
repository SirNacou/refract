package shortenurl

import (
	"context"
	"time"

	"github.com/SirNacou/refract/api/internal/domain"
	"github.com/danielgtaylor/huma/v2"
)

type Command struct {
	OriginalURL string
	Domain      string
	UserID      string
	ExpiresAt   *time.Time
}

type CommandResponse struct {
	ShortCode string
	Domain    string
}

type CommandHandler struct {
	repo          domain.URLRepository
	defaultDomain string
}

func NewCommandHandler(repo domain.URLRepository, defaultDomain string) *CommandHandler {
	return &CommandHandler{
		repo:          repo,
		defaultDomain: defaultDomain,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*CommandResponse, error) {
	domainName := ""
	if cmd.Domain != "" {
	}

	url := domain.NewURL(cmd.OriginalURL, "", domainName, "", "", cmd.UserID, cmd.ExpiresAt)
	err := h.repo.Create(ctx, url)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to shorten URL", err)
	}

	return &CommandResponse{
		ShortCode: url.ShortCode,
		Domain:    url.Domain,
	}, nil
}
