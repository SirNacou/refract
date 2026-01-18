package postgres

import (
	"github.com/SirNacou/refract/services/api/internal/domain/url"
)

type PostgresURLRepository struct {
	db *DBConnection
}

func NewPostgresURLRepository(db *DBConnection) *PostgresURLRepository {
	return &PostgresURLRepository{db: db}
}

func (r *PostgresURLRepository) Create(u *url.URL) error {
	// TODO: Implement database insert
	panic("not implemented")
}

func (r *PostgresURLRepository) GetBySnowflakeID(snowflakeID uint64) (*url.URL, error) {
	// TODO: Implement database query
	panic("not implemented")
}

func (r *PostgresURLRepository) GetByCustomAlias(alias string) (*url.URL, error) {
	// TODO: Implement database query
	panic("not implemented")
}

func (r *PostgresURLRepository) GetByCreatorID(userID string) ([]url.URL, error) {
	// TODO: Implement database query
	panic("not implemented")
}
