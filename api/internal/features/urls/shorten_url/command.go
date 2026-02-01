package shortenurl

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/SirNacou/refract/api/internal/domain"
	"github.com/danielgtaylor/huma/v2"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

type Command struct {
	OriginalURL string
	UserID      string
	ExpiresAt   *time.Time
}

type CommandResponse struct {
	ShortURL string
}

type CommandHandler struct {
	repo           domain.URLRepository
	valkey         valkeyaside.CacheAsideClient
	defaultBaseURL string
	redirectKey    string
}

func NewCommandHandler(repo domain.URLRepository, valkey valkeyaside.CacheAsideClient, defaultBaseURL, redirectKey string) *CommandHandler {
	return &CommandHandler{
		repo:           repo,
		valkey:         valkey,
		defaultBaseURL: defaultBaseURL,
		redirectKey:    redirectKey,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*CommandResponse, error) {
	u := domain.NewURL(cmd.OriginalURL, "", "", cmd.UserID, domain.NewShortCode(""), cmd.ExpiresAt)
	err := h.repo.Create(ctx, u)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to shorten URL", err)
	}

	expiration := time.Hour * 24 * 365
	if u.ExpiresAt != nil {
		exp := time.Until(*u.ExpiresAt)
		if exp < expiration {
			expiration = exp
		}
	}

	key := strings.Replace(h.redirectKey, "{short_code}", u.ShortCode.String(), 1)
	err = h.valkey.Client().
		Do(ctx,
			h.valkey.Client().
				B().
				Set().
				Key(key).
				Value(u.OriginalURL).
				Ex(expiration).
				Build()).
		Error()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to cache short code", "short_code", u.ShortCode)
	}

	shortURL, err := url.JoinPath(h.defaultBaseURL, u.ShortCode.String())
	if err != nil {
		return nil, err
	}

	return &CommandResponse{
		ShortURL: shortURL,
	}, nil
}
