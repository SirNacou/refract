package domain

import "time"

type URL struct {
	ID          uint64
	OriginalURL string
	ShortCode   string
	Title       string
	Notes       string
	UserID      string
	ExpiryDate  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type URLRepository interface {
	ListByUser(userID string) ([]URL, error)
}
