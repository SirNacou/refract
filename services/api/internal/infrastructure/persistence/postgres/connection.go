package postgres

import (
	"context"

	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBConnection wraps a pgxpool.Pool to provide database connection pooling
type DBConnection struct {
	Pool *pgxpool.Pool
}

func NewDBConnection(ctx context.Context, cfg *config.DatabaseConfig) (*DBConnection, error) {
	dsn := cfg.GetDatabaseDSN()
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Apply pool size configuration
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	return &DBConnection{Pool: pool}, nil
}
