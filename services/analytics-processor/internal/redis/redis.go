package redis

import (
	"github.com/SirNacou/refract/services/analytics-processor/internal/config"
	"github.com/valkey-io/valkey-go"
)

func NewValkeyClient(cfg *config.Config) (valkey.Client, error) {
	return valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.GetRedisAddress()},
		Password:    cfg.REDIS_PASSWORD,
	})
}
