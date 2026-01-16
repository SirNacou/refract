package url

type URLRepository interface {
	Create(url *URL) error
	GetBySnowflakeID(snowflakeID uint64) (*URL, error)
	GetByCustomAlias(alias string) (*URL, error)
	GetByCreatorID(userID string) ([]URL, error)
}
