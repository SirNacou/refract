package cache

import (
	"fmt"

	"github.com/valkey-io/valkey-go"
)

type RedisCache struct {
	client valkey.Client
}

func NewRedisCache(host string, port int) (*RedisCache, error) {
	config := valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", host, port)},
	}

	client, err := valkey.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &RedisCache{client}, nil
}

func (rc *RedisCache) Client() valkey.Client {
	return rc.client
}

func (rc *RedisCache) Close() {
	rc.client.Close()
}
