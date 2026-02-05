package domain

import (
	"context"
	"time"
)

type Status = string

const (
	Active   Status = "active"
	Expired  Status = "expired"
	Disabled Status = "disabled"
)

type URL struct {
	ID          SnowflakeID
	OriginalURL string
	ShortCode   ShortCode
	Title       string
	Notes       string
	UserID      string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Status      Status
}

func NewURL(originalURL, title, notes, userID string, shortCode *ShortCode, expiresAt *time.Time) *URL {
	id := NewSnowflakeID()
	if shortCode == nil {
		sc := GenerateShortcode(id)
		shortCode = &sc
	}
	return &URL{
		ID:          id,
		OriginalURL: originalURL,
		ShortCode:   *shortCode,
		Title:       title,
		Notes:       notes,
		UserID:      userID,
		ExpiresAt:   expiresAt,
		Status:      Active,
	}
}

type URLRepository interface {
	ListByUser(ctx context.Context, userID string) ([]URL, error)
	GetActiveURLByShortCode(ctx context.Context, shortCode ShortCode) (*URL, error)
	Create(ctx context.Context, url *URL) error
	CountByUser(ctx context.Context, userID string) (int64, error)
	CountActiveByUser(ctx context.Context, userID string) (int64, error)
}
