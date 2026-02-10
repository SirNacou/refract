package shortenurl

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/SirNacou/refract/api/internal/domain"
	"github.com/SirNacou/refract/api/internal/infrastructure/validator"
	"github.com/danielgtaylor/huma/v2"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

type Command struct {
	Title       string     `validate:"required,max=255"`
	OriginalURL string     `validate:"required,url,max=2048"`
	UserID      string     `validate:"required"`
	ExpiresAt   *time.Time `validate:"omitempty"`
}

type CommandResponse struct {
	ShortURL string
}

type CommandHandler struct {
	repo           domain.URLRepository
	valkey         valkeyaside.CacheAsideClient
	chConn         clickhouse.Conn
	defaultBaseURL string
	redirectKey    string
}

func NewCommandHandler(repo domain.URLRepository, valkey valkeyaside.CacheAsideClient, chConn clickhouse.Conn, defaultBaseURL, redirectKey string) *CommandHandler {
	return &CommandHandler{
		repo:           repo,
		valkey:         valkey,
		chConn:         chConn,
		defaultBaseURL: defaultBaseURL,
		redirectKey:    redirectKey,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*CommandResponse, error) {
	err := validator.GetValidator().StructCtx(ctx, cmd)
	if err != nil {
		return nil, err
	}

	u := domain.NewURL(cmd.OriginalURL, cmd.Title, "", cmd.UserID, domain.NewShortCode(""), cmd.ExpiresAt)
	err = h.repo.Create(ctx, u)
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

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		key := strings.Replace(h.redirectKey, "{short_code}", u.ShortCode.String(), 1)
		err := h.valkey.Client().
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
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err := h.chConn.Exec(ctx, `
	INSERT INTO refract.urls (short_code, original_url, title, created_by) VALUES (
		?, ?, ?, ?
	)
	`, u.ShortCode.String(), u.OriginalURL, u.Title, u.UserID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to insert URL into ClickHouse", "error", err)
		}
	}()

	shortURL := strings.Join([]string{h.defaultBaseURL, u.ShortCode.String()}, "/")

	return &CommandResponse{
		ShortURL: shortURL,
	}, nil
}
