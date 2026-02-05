package persistence

import (
	"context"

	"github.com/SirNacou/refract/api/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool    *pgxpool.Pool
	Querier db.Querier
}

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	p, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	err = p.Ping(ctx)
	if err != nil {
		return nil, err
	}

	querier := db.New(p)

	return &DB{Pool: p, Querier: querier}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}
