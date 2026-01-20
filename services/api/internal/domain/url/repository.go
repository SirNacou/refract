package url

import "context"

type URLRepository interface {
	Create(ctx context.Context, url *URL) error
	GetBySnowflakeID(ctx context.Context, snowflakeID uint64) (*URL, error)
	GetByCustomAlias(ctx context.Context, alias string) (*URL, error)
	GetByCreatorID(ctx context.Context, userID string) ([]URL, error)
}
