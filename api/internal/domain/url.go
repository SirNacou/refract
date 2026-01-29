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
	Domain      string
	Title       string
	Notes       string
	UserID      string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Status      Status
}

func NewURL(originalURL, domain, title, notes, userID string, shortCode *ShortCode, expiresAt *time.Time) *URL {
	id := NewSnowflakeID()
	if shortCode == nil {
		sc := GenerateShortcode(id)
		shortCode = &sc
	}
	return &URL{
		ID:          id,
		OriginalURL: originalURL,
		ShortCode:   *shortCode,
		Domain:      domain,
		Title:       title,
		Notes:       notes,
		UserID:      userID,
		ExpiresAt:   expiresAt,
		Status:      Active,
	}
}

type URLRepository interface {
	ListByUser(ctx context.Context, userID string) ([]URL, error)
	Create(ctx context.Context, url *URL) error
}
