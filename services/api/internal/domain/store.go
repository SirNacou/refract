package domain

import (
	"context"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
)

type Store interface {
	URLs() url.URLRepository

	ExecTx(ctx context.Context, fn func(Store) error) error
}
