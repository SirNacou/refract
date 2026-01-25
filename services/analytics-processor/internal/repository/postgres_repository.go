package repository

import (
	"context"

	"github.com/SirNacou/refract/services/analytics-processor/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var columns = []string{
	"time", "event_id", "url_id", "referrer", "user_agent", "visitor_hash",
	"country_code", "country_name", "city", "latitude", "longitude",
	"device_type", "browser", "operating_system",
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) (*PostgresRepository, error) {
	return &PostgresRepository{pool: pool}, nil
}

func (r *PostgresRepository) InsertClickEvents(ctx context.Context, events []domain.ClickEvent) error {
	rows := [][]any{}
	for _, e := range events {
		rows = append(rows, []any{
			e.Timestamp,
			e.EventID,
			e.URLID,
			e.Referrer,
			e.UserAgent,
			e.VisitorHash,
			e.CountryCode,
			e.CountryName,
			e.City,
			e.Latitude,
			e.Longitude,
			e.DeviceType,
			e.Browser,
			e.OperatingSystem,
		})
	}

	_, err := r.pool.CopyFrom(ctx,
		pgx.Identifier{"click_events"},
		columns,
		pgx.CopyFromRows(rows))
	if err != nil {
		return err
	}

	return nil
}
