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
	ShortCode   string
	Domain      string
	Title       string
	Notes       string
	UserID      string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Status      Status
}

func NewURL(originalURL, shortCode, domain, title, notes, userID string, expiresAt *time.Time) *URL {
	return &URL{
		ID:          NewSnowflakeID(),
		OriginalURL: originalURL,
		ShortCode:   shortCode,
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
