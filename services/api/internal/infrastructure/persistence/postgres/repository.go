package postgres

import (
	"context"
	"fmt"

	"github.com/SirNacou/refract/services/api/internal/domain"
	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/persistence/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	*db.Queries
	connPool *pgxpool.Pool
}

func NewSQLRepository(pool *pgxpool.Pool) domain.Store {
	return &SQLRepository{
		Queries:  db.New(pool),
		connPool: pool,
	}
}

// ExecTx implements [domain.Store].
func (s *SQLRepository) ExecTx(ctx context.Context, fn func(domain.Store) error) error {
	tx, err := s.connPool.Begin(ctx)

	if err != nil {
		return err
	}

	txRepo := &SQLRepository{
		Queries:  db.New(tx),
		connPool: s.connPool,
	}

	err = fn(txRepo)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

// URLs implements [domain.Store].
func (s *SQLRepository) URLs() url.URLRepository {
	return NewPostgresURLRepository(s.Queries)
}
