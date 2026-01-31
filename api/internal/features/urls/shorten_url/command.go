package shortenurl

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/SirNacou/refract/api/internal/domain"
	"github.com/danielgtaylor/huma/v2"
	"github.com/valkey-io/valkey-go/valkeyaside"
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
	valkey        valkeyaside.CacheAsideClient
	defaultDomain string
	redirectKey   string
}

func NewCommandHandler(repo domain.URLRepository, valkey valkeyaside.CacheAsideClient, defaultDomain, redirectKey string) *CommandHandler {
	return &CommandHandler{
		repo:          repo,
		valkey:        valkey,
		defaultDomain: defaultDomain,
		redirectKey:   redirectKey,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*CommandResponse, error) {
	domainName := ""
	if cmd.Domain != "" {
		domainName = cmd.Domain
	}

	url := domain.NewURL(cmd.OriginalURL, domainName, "", "", cmd.UserID, domain.NewShortCode(""), cmd.ExpiresAt)
	err := h.repo.Create(ctx, url)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to shorten URL", err)
	}

	expiration := time.Hour * 24 * 365
	if url.ExpiresAt != nil {
		exp := time.Until(*url.ExpiresAt)
		if exp < expiration {
			expiration = exp
		}
	}

	key := strings.Replace(h.redirectKey, "{short_code}", url.ShortCode.String(), 1)
	err = h.valkey.Client().
		Do(ctx,
			h.valkey.Client().
				B().
				Set().
				Key(key).
				Value(url.OriginalURL).
				Ex(expiration).
				Build()).
		Error()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to cache short code", "short_code", url.ShortCode)
	}

	return &CommandResponse{
		ShortCode: url.ShortCode.String(),
		Domain:    url.Domain,
	}, nil
}
