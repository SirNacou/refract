package models

import "time"

type URL struct {
	ID          int64          `json:"id"`
	ShortCode   string         `json:"short_code"`
	OriginalURL string         `json:"original_url"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	ClickCount  int64          `json:"click_count"`
	IsActive    bool           `json:"is_active"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
