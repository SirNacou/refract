package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/SirNacou/refract/services/api/internal/application/service"
	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/errors"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
)

type RateLimiter struct {
	cache    service.Cache
	fallback *sync.Map
	config   *config.SecurityConfig
}

type memoryCounter struct {
	count     int
	resetTime time.Time
	mu        sync.Mutex
}

func NewRateLimiter(cache service.Cache, cfg *config.SecurityConfig) *RateLimiter {
	return &RateLimiter{
		cache:    cache,
		fallback: &sync.Map{},
		config:   cfg,
	}
}

func (rl *RateLimiter) RateLimitPerUser() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := authorization.UserID(r.Context())

			remaining, resetTime, err := rl.checkLimit(r.Context(), userID)

			if err == ErrRateLimitExceeded {
				rl.writeRateLimitError(w, r, resetTime)
				return
			}

			rl.addRateLimitHeaders(w, remaining, resetTime)

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) addRateLimitHeaders(w http.ResponseWriter, remaining int, resetTime int64) {
	limit := rl.config.RateLimitPerUser
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
}

func (rl *RateLimiter) writeRateLimitError(w http.ResponseWriter, r *http.Request, resetTime int64) {
	limit := rl.config.RateLimitPerUser

	// Set rate limit headers BEFORE WriteRateLimitExceeded (which calls WriteHeader)
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

	errors.WriteRateLimitExceeded(w, r, fmt.Sprintf("Rate limit exceeded (%d requests per hour)", limit))
}

func (rl *RateLimiter) checkLimit(ctx context.Context, userID string) (
	remaining int,
	resettime int64,
	err error,
) {
	limit := rl.config.RateLimitPerUser
	window := rl.config.RateLimitWindow

	var count int
	if rl.cache != nil {
		count, err = rl.checkRedis(ctx, userID, window)
		if err != nil {
			slog.WarnContext(ctx, "Redis unavailable, using in-memory rate limiter",
				"user_id", userID,
				"error", err)
			count = rl.checkInMemory(userID, window)
		}
	} else {
		// Redis not configured, use in-memory fallback
		count = rl.checkInMemory(userID, window)
	}

	remaining = limit - count
	if remaining < 0 {
		remaining = 0
	}

	resettime = time.Now().Add(window).Unix()

	if count > limit {
		return remaining, resettime, ErrRateLimitExceeded
	}

	return remaining, resettime, nil
}

func (rl *RateLimiter) checkRedis(ctx context.Context, userID string, window time.Duration) (int, error) {
	key := fmt.Sprintf("ratelimit:user:%s", userID)

	incrCmd := rl.cache.Client().B().Incr().Key(key).Build()
	count, err := rl.cache.Client().Do(ctx, incrCmd).AsInt64()
	if err != nil {
		return 0, fmt.Errorf("redis INCR failed: %w", err)
	}

	if count == 1 {
		expireCmd := rl.cache.Client().B().Expire().Key(key).Seconds(int64(window.Seconds())).Build()
		rl.cache.Client().Do(ctx, expireCmd)
	}

	return int(count), nil
}

func (rl *RateLimiter) checkInMemory(userID string, window time.Duration) int {
	key := fmt.Sprintf("user:%s", userID)
	now := time.Now()

	val, _ := rl.fallback.LoadOrStore(key, &memoryCounter{
		count:     0,
		resetTime: now.Add(window),
	})

	counter := val.(*memoryCounter)
	counter.mu.Lock()
	defer counter.mu.Unlock()

	if now.After(counter.resetTime) {
		counter.count = 1
		counter.resetTime = now.Add(window)
		return 1
	}

	counter.count++
	return counter.count
}
