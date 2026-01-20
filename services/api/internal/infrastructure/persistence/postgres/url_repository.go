package postgres

import (
	"context"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/persistence/db"
)

type PostgresURLRepository struct {
	db db.Querier
}

func NewPostgresURLRepository(db db.Querier) url.URLRepository {
	return &PostgresURLRepository{db: db}
}

// Create implements [url.URLRepository].
func (p *PostgresURLRepository) Create(ctx context.Context, url *url.URL) error {
	code := url.ShortCode.String()
	p.db.CreateURL(ctx, db.CreateURLParams{
		SnowflakeID:    int64(url.ID),
		ShortCode:      code,
		DestinationUrl: url.DestinationURL,
		Title:          url.Title,
		Notes:          &url.Notes,
		Status:         string(url.Status),
		CreatedAt:      url.CreatedAt,
		UpdatedAt:      url.UpdatedAt,
		ExpiresAt:      url.ExpiresAt,
		CreatorUserID:  url.CreatorUserID,
	})
	return nil
}

// GetByCreatorID implements [url.URLRepository].
func (p *PostgresURLRepository) GetByCreatorID(ctx context.Context, userID string) ([]url.URL, error) {
	panic("unimplemented")
}

// GetByCustomAlias implements [url.URLRepository].
func (p *PostgresURLRepository) GetByCustomAlias(ctx context.Context, shortCode string) (*url.URL, error) {
	r, err := p.db.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	s, err := url.ParseShortCode(r.ShortCode)
	if err != nil {
		return nil, err
	}

	return &url.URL{
		ID:             uint64(r.SnowflakeID),
		ShortCode:      *s,
		DestinationURL: r.DestinationUrl,
		Title:          r.Title,
		Notes:          *r.Notes,
		Status:         url.Status(r.Status),
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		ExpiresAt:      r.ExpiresAt,
		CreatorUserID:  r.CreatorUserID,
		TotalClicks:    uint64(r.TotalClicks),
		LastClickedAt:  r.LastClickedAt,
	}, nil
}

// GetBySnowflakeID implements [url.URLRepository].
func (p *PostgresURLRepository) GetBySnowflakeID(ctx context.Context, snowflakeID uint64) (*url.URL, error) {
	panic("unimplemented")
}
