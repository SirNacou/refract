package persistence

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/cache"
)

// CachedURLRepository wraps a URL repository with caching
type CachedURLRepository struct {
	repo   url.Repository
	cache  cache.Cache
	logger *slog.Logger
}

// NewCachedURLRepository creates a cached repository decorator
func NewCachedURLRepository(repo url.Repository, cache cache.Cache, logger *slog.Logger) *CachedURLRepository {
	return &CachedURLRepository{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// cacheKey generates the cache key for a short code
func (r *CachedURLRepository) cacheKey(shortCode string) string {
	return fmt.Sprintf("url:redirect:%s", shortCode)
}

// calculateTTL computes smart TTL: min(expiresAt - now, 24h)
func (r *CachedURLRepository) calculateTTL(expiresAt time.Time) time.Duration {
	untilExpiration := time.Until(expiresAt)
	maxTTL := 24 * time.Hour

	if untilExpiration < maxTTL {
		return untilExpiration
	}
	return maxTTL
}

// Save persists URL to database and caches it (write-through)
func (r *CachedURLRepository) Save(ctx context.Context, urlEntity *url.URL) error {
	// Save to database first
	if err := r.repo.Save(ctx, urlEntity); err != nil {
		return err
	}

	// Cache the redirect mapping (best effort - don't fail if cache fails)
	ttl := r.calculateTTL(urlEntity.ExpiresAt())
	cacheKey := r.cacheKey(urlEntity.ShortCode().String())
	cacheValue := urlEntity.OriginalURL().String()

	if err := r.cache.Set(ctx, cacheKey, cacheValue, ttl); err != nil {
		// Log warning but don't fail the operation
		r.logger.Warn("Failed to cache URL on creation",
			slog.String("short_code", urlEntity.ShortCode().String()),
			slog.Duration("ttl", ttl),
			slog.String("error", err.Error()),
		)
	} else {
		r.logger.Info("URL cached on creation",
			slog.String("short_code", urlEntity.ShortCode().String()),
			slog.String("original_url", cacheValue),
			slog.Duration("ttl", ttl),
		)
	}

	return nil
}

// FindByShortCode - NO CACHING (management service returns full entity)
func (r *CachedURLRepository) FindByShortCode(ctx context.Context, code url.ShortCode) (*url.URL, error) {
	return r.repo.FindByShortCode(ctx, code)
}

// ExistsByShortCode - delegate to wrapped repository
func (r *CachedURLRepository) ExistsByShortCode(ctx context.Context, code url.ShortCode) (bool, error) {
	return r.repo.ExistsByShortCode(ctx, code)
}

// UpdateExpiration - delegate to wrapped repository
func (r *CachedURLRepository) UpdateExpiration(ctx context.Context, code url.ShortCode, expiresAt time.Time) error {
	return r.repo.UpdateExpiration(ctx, code, expiresAt)
}

// IncrementClickCount - delegate to wrapped repository
func (r *CachedURLRepository) IncrementClickCount(ctx context.Context, code url.ShortCode) error {
	return r.repo.IncrementClickCount(ctx, code)
}
