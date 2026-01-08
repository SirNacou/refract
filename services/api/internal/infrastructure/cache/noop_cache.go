package cache

import (
	"context"
	"time"
)

// NoopCache is a no-op cache implementation
type NoopCache struct{}

// NewNoopCache creates a new no-op cache
func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

func (c *NoopCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

func (c *NoopCache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (c *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoopCache) Ping(ctx context.Context) error {
	return nil
}

func (c *NoopCache) Close() error {
	return nil
}
