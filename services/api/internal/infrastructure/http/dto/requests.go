package dto

import "time"

// CreateURLRequest represents the HTTP request to create a shortened URL
type CreateURLRequest struct {
	OriginalURL string     `json:"original_url"`
	Domain      string     `json:"domain"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}
