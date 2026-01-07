package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/persistence/postgres/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresURLRepository implements the url.Repository interface using PostgreSQL
type PostgresURLRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewPostgresURLRepository creates a new PostgreSQL-based URL repository
func NewPostgresURLRepository(pool *pgxpool.Pool) *PostgresURLRepository {
	return &PostgresURLRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a URL entity
func (r *PostgresURLRepository) Save(ctx context.Context, urlEntity *url.URL) error {
	// Marshal metadata to JSON
	metadataJSON, err := json.Marshal(urlEntity.Metadata())
	if err != nil {
		return url.NewInternalError("SERIALIZATION_ERROR", "Failed to serialize metadata", err)
	}

	err = r.queries.CreateURL(ctx, generated.CreateURLParams{
		ID:                 urlEntity.ID(),
		ShortCode:          urlEntity.ShortCode().String(),
		OriginalUrl:        urlEntity.OriginalURL().String(),
		Domain:             urlEntity.Domain().String(),
		ExpiresAt:          timeToTimestamptz(urlEntity.ExpiresAt()),
		HasFixedExpiration: urlEntity.HasFixedExpiration(),
		IsActive:           urlEntity.IsActive(),
		Metadata:           metadataJSON,
	})

	if err != nil {
		// Check for unique constraint violation (duplicate short code)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return url.NewConflictError(
				"DUPLICATE_SHORT_CODE",
				fmt.Sprintf("Short code '%s' already exists", urlEntity.ShortCode().String()),
			)
		}
		return url.NewInternalError("DB_ERROR", "Failed to save URL", err)
	}

	return nil
}

// FindByShortCode retrieves a URL by its short code
func (r *PostgresURLRepository) FindByShortCode(ctx context.Context, code url.ShortCode) (*url.URL, error) {
	row, err := r.queries.GetURLByShortCode(ctx, code.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, url.NewNotFoundError("URL_NOT_FOUND", "URL not found")
		}
		return nil, url.NewInternalError("DB_ERROR", "Failed to query URL", err)
	}

	return rowToEntity(row)
}

// ExistsByShortCode checks if a short code already exists
func (r *PostgresURLRepository) ExistsByShortCode(ctx context.Context, code url.ShortCode) (bool, error) {
	exists, err := r.queries.ExistsByShortCode(ctx, code.String())
	if err != nil {
		return false, url.NewInternalError("DB_ERROR", "Failed to check short code existence", err)
	}
	return exists, nil
}

// UpdateExpiration updates the expiration time for a URL
func (r *PostgresURLRepository) UpdateExpiration(ctx context.Context, code url.ShortCode, expiresAt time.Time) error {
	err := r.queries.UpdateExpiration(ctx, generated.UpdateExpirationParams{
		ExpiresAt: timeToTimestamptz(expiresAt),
		ShortCode: code.String(),
	})
	if err != nil {
		return url.NewInternalError("DB_ERROR", "Failed to update expiration", err)
	}
	return nil
}

// IncrementClickCount increments the click count for a URL
func (r *PostgresURLRepository) IncrementClickCount(ctx context.Context, code url.ShortCode) error {
	err := r.queries.IncrementClickCount(ctx, code.String())
	if err != nil {
		return url.NewInternalError("DB_ERROR", "Failed to increment click count", err)
	}
	return nil
}

// rowToEntity converts a database row to a domain entity
func rowToEntity(row generated.GetURLByShortCodeRow) (*url.URL, error) {
	// Parse value objects
	shortCode, err := url.NewShortCode(row.ShortCode)
	if err != nil {
		return nil, fmt.Errorf("invalid short code in database: %w", err)
	}

	originalURL, err := url.NewOriginalURL(row.OriginalUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid original URL in database: %w", err)
	}

	domain, err := url.NewDomain(row.Domain)
	if err != nil {
		return nil, fmt.Errorf("invalid domain in database: %w", err)
	}

	// Unmarshal metadata
	var metadata map[string]any
	if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Reconstruct entity
	return url.Reconstruct(
		row.ID,
		shortCode,
		originalURL,
		domain,
		timestamptzToTime(row.CreatedAt),
		timestamptzToTime(row.UpdatedAt),
		timestamptzToTime(row.ExpiresAt),
		row.HasFixedExpiration,
		row.ClickCount,
		row.IsActive,
		metadata,
	), nil
}

// Helper functions for pgtype.Timestamptz conversions
func timeToTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

func timestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}
