package cache

import (
	"fmt"

	"github.com/SirNacou/refract/services/api/internal/application/service"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

type RedisCache struct {
	valkeyaside.CacheAsideClient
}

func NewRedisCache(host string, port int) (service.Cache, error) {
	config := valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", host, port)},
	}

	c, err := valkeyaside.NewClient(valkeyaside.ClientOption{
		ClientOption: config,
	})

	if err != nil {
		return nil, err
	}

	return c, nil
}
