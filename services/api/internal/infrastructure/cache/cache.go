// Package cache provides caching operations for the application
package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Set stores a key-value pair with TTL
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Get retrieves a value by key
	Get(ctx context.Context, key string) (string, error)

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error

	// Ping checks if cache is available
	Ping(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}
