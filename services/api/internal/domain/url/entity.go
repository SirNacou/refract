package url

import "time"

// URL is the aggregate root for the URL shortener domain
type URL struct {
	id                 int64
	shortCode          ShortCode
	originalURL        OriginalURL
	domain             Domain
	createdAt          time.Time
	updatedAt          time.Time
	expiresAt          time.Time
	hasFixedExpiration bool
	clickCount         int64
	isActive           bool
	metadata           map[string]any
}

// NewURLWithFixedExpiration creates a URL with a user-specified expiration (never renews)
func NewURLWithFixedExpiration(
	id int64,
	shortCode ShortCode,
	originalURL OriginalURL,
	domain Domain,
	expiresAt time.Time,
) (*URL, error) {
	now := time.Now()

	// Validate expiration is not in the past
	if expiresAt.Before(now) {
		return nil, NewValidationError(
			"INVALID_EXPIRATION",
			"Expiration date cannot be in the past",
		)
	}

	// Validate expiration is not too far in the future (10 years)
	maxExpiration := now.AddDate(10, 0, 0)
	if expiresAt.After(maxExpiration) {
		return nil, NewValidationError(
			"INVALID_EXPIRATION",
			"Expiration date cannot be more than 10 years in the future",
		)
	}

	return &URL{
		id:                 id,
		shortCode:          shortCode,
		originalURL:        originalURL,
		domain:             domain,
		createdAt:          now,
		updatedAt:          now,
		expiresAt:          expiresAt,
		hasFixedExpiration: true,
		clickCount:         0,
		isActive:           true,
		metadata:           make(map[string]any),
	}, nil
}

// NewURLWithActivityExpiration creates a URL that renews on every click (12 months from now)
func NewURLWithActivityExpiration(
	id int64,
	shortCode ShortCode,
	originalURL OriginalURL,
	domain Domain,
) *URL {
	now := time.Now()
	expiresAt := now.AddDate(0, 12, 0) // 12 months from now

	return &URL{
		id:                 id,
		shortCode:          shortCode,
		originalURL:        originalURL,
		domain:             domain,
		createdAt:          now,
		updatedAt:          now,
		expiresAt:          expiresAt,
		hasFixedExpiration: false,
		clickCount:         0,
		isActive:           true,
		metadata:           make(map[string]any),
	}
}

// Reconstruct creates a URL from persistence (for repository)
func Reconstruct(
	id int64,
	shortCode ShortCode,
	originalURL OriginalURL,
	domain Domain,
	createdAt time.Time,
	updatedAt time.Time,
	expiresAt time.Time,
	hasFixedExpiration bool,
	clickCount int64,
	isActive bool,
	metadata map[string]any,
) *URL {
	if metadata == nil {
		metadata = make(map[string]any)
	}

	return &URL{
		id:                 id,
		shortCode:          shortCode,
		originalURL:        originalURL,
		domain:             domain,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
		expiresAt:          expiresAt,
		hasFixedExpiration: hasFixedExpiration,
		clickCount:         clickCount,
		isActive:           isActive,
		metadata:           metadata,
	}
}

// Business Methods

// IsExpired checks if the URL has expired
func (u *URL) IsExpired() bool {
	return time.Now().After(u.expiresAt)
}

// ShouldRenewOnClick returns true if this URL renews expiration on clicks
func (u *URL) ShouldRenewOnClick() bool {
	return !u.hasFixedExpiration
}

// RenewExpiry resets the expiration to 12 months from now
// Only works for activity-based URLs
func (u *URL) RenewExpiry() {
	if u.ShouldRenewOnClick() {
		u.expiresAt = time.Now().AddDate(0, 12, 0)
		u.updatedAt = time.Now()
	}
}

// IncrementClickCount increments the click counter
func (u *URL) IncrementClickCount() {
	u.clickCount++
	u.updatedAt = time.Now()
}

// Deactivate marks the URL as inactive
func (u *URL) Deactivate() {
	u.isActive = false
	u.updatedAt = time.Now()
}

// Getters (for read-only access)

func (u *URL) ID() int64 {
	return u.id
}

func (u *URL) ShortCode() ShortCode {
	return u.shortCode
}

func (u *URL) OriginalURL() OriginalURL {
	return u.originalURL
}

func (u *URL) Domain() Domain {
	return u.domain
}

func (u *URL) CreatedAt() time.Time {
	return u.createdAt
}

func (u *URL) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *URL) ExpiresAt() time.Time {
	return u.expiresAt
}

func (u *URL) HasFixedExpiration() bool {
	return u.hasFixedExpiration
}

func (u *URL) ClickCount() int64 {
	return u.clickCount
}

func (u *URL) IsActive() bool {
	return u.isActive
}

func (u *URL) Metadata() map[string]any {
	// Return a copy to prevent external modification
	meta := make(map[string]any, len(u.metadata))
	for k, v := range u.metadata {
		meta[k] = v
	}
	return meta
}
