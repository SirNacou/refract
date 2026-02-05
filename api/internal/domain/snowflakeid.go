package domain

import (
	"github.com/SirNacou/refract/api/internal/infrastructure/snowflake"
)

type SnowflakeID uint64

func NewSnowflakeID() SnowflakeID {
	return SnowflakeID(snowflake.Generator.Generate())
}

func (id SnowflakeID) Int64() int64 {
	return int64(id)
}