package domain

import (
	"github.com/SirNacou/refract/api/internal/infrastructure/snowflake"
)

type SnowflakeID uint64

func NewSnowflakeID() SnowflakeID {
	return SnowflakeID(snowflake.Generator.Generate())
}
