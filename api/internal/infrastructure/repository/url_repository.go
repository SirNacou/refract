package repository

import (
	"context"
	"log/slog"

	"github.com/SirNacou/refract/api/internal/db"
	"github.com/SirNacou/refract/api/internal/domain"
)

type PostgresURLRepository struct {
	querier db.Querier
}

func NewPostgresURLRepository(querier db.Querier) domain.URLRepository {
	return &PostgresURLRepository{
		querier: querier,
	}
}

// ListByUser implements [domain.URLRepository].
func (p *PostgresURLRepository) ListByUser(ctx context.Context, userID string) ([]domain.URL, error) {
	urls, err := p.querier.ListURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(urls) == 0 {
		slog.Info("returning empty array of URLs for user", "userID", userID)
	}

	result := make([]domain.URL, 0, len(urls))
	for _, u := range urls {
		result = append(result, *toDomainURL(&u))
	}

	return result, nil
}

// FirstByShortCode implements [domain.URLRepository].
func (p *PostgresURLRepository) GetURLByShortCode(ctx context.Context, shortCode domain.ShortCode) (*domain.URL, error) {
	url, err := p.querier.GetURLByShortCode(ctx, shortCode.String())
	if err != nil {
		return nil, err
	}

	return toDomainURL(&url), nil
}

// Create implements [domain.URLRepository].
func (p *PostgresURLRepository) Create(ctx context.Context, url *domain.URL) error {
	_, err := p.querier.CreateURL(ctx, db.CreateURLParams{
		ID:          url.ID.Int64(),
		ShortCode:   url.ShortCode.String(),
		OriginalUrl: url.OriginalURL,
		UserID:      url.UserID,
		Domain:      url.Domain,
		ExpiresAt:   url.ExpiresAt,
	})
	if err != nil {
		return err
	}

	return nil
}

func toDomainURL(u *db.Url) *domain.URL {
	return &domain.URL{
		ID:          domain.SnowflakeID(u.ID),
		OriginalURL: u.OriginalUrl,
		ShortCode:   domain.ShortCode(u.ShortCode),
		Domain:      u.Domain,
		UserID:      u.UserID,
		ExpiresAt:   u.ExpiresAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Status:      u.Status,
		Title:       "",
		Notes:       "",
	}
}
