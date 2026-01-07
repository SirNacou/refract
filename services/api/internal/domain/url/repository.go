package url

import (
	"context"
	"time"
)

// Repository defines the interface for URL persistence
type Repository interface {
	// Save persists a URL entity
	Save(ctx context.Context, url *URL) error

	// FindByShortCode retrieves a URL by its short code
	FindByShortCode(ctx context.Context, code ShortCode) (*URL, error)

	// ExistsByShortCode checks if a short code already exists
	ExistsByShortCode(ctx context.Context, code ShortCode) (bool, error)

	// UpdateExpiration updates the expiration time for a URL
	UpdateExpiration(ctx context.Context, code ShortCode, expiresAt time.Time) error

	// IncrementClickCount increments the click count for a URL
	IncrementClickCount(ctx context.Context, code ShortCode) error
}
