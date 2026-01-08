package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// ValkeyCache implements Cache using go-redis (compatible with Valkey/Redis)
type ValkeyCache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewValkeyCache creates a new Valkey/Redis cache client using go-redis
func NewValkeyCache(host string, port int, password string, db int, logger *slog.Logger) (*ValkeyCache, error) {
	// Create Redis client (works with Valkey too)
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	return &ValkeyCache{
		client: client,
		logger: logger,
	}, nil
}

func (c *ValkeyCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := c.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		c.logger.Error("Failed to set cache key",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("cache set failed: %w", err)
	}

	c.logger.Debug("Cache key set",
		slog.String("key", key),
		slog.Duration("ttl", ttl),
	)
	return nil
}

func (c *ValkeyCache) Get(ctx context.Context, key string) (string, error) {
	value, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key doesn't exist - not an error
		c.logger.Debug("Cache miss", slog.String("key", key))
		return "", nil
	}
	if err != nil {
		c.logger.Warn("Failed to get cache key",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("cache get failed: %w", err)
	}

	c.logger.Debug("Cache hit", slog.String("key", key))
	return value, nil
}

func (c *ValkeyCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete cache key",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("cache delete failed: %w", err)
	}
	return nil
}

func (c *ValkeyCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *ValkeyCache) Close() error {
	return c.client.Close()
}
