package postgres

import (
	"context"

	"github.com/SirNacou/refract/services/api/internal/domain"
	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/persistence/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLStore struct {
	conn *pgxpool.Pool
	q    db.Querier
}

func NewSQLStore(pool *pgxpool.Pool) domain.Store {
	return &SQLStore{
		conn: pool,
		q:    db.New(pool),
	}
}

// 2. Implement the Transaction Logic
func (s *SQLStore) ExecTx(ctx context.Context, fn func(domain.Store) error) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}

	txStore := &SQLStore{
		conn: s.conn,
		q:    db.New(tx), // Bind ALL domains to this ONE transaction
	}

	err = fn(txStore)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// URLs implements [domain.Store].
func (s *SQLStore) URLs() url.URLRepository {
	panic("unimplemented")
}
