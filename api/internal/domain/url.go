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

type SnowflakeID uint64

type URL struct {
	ID          SnowflakeID
	OriginalURL string
	ShortCode   string
	Title       string
	Notes       string
	UserID      string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Status      Status
}

type URLRepository interface {
	ListByUser(ctx context.Context, userID string) ([]URL, error)
}
