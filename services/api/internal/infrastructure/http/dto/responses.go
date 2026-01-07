package dto

import "time"

// ErrorResponse represents an HTTP error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CreateURLResponse represents the response after creating a URL
type CreateURLResponse struct {
	ShortCode          string    `json:"short_code"`
	ShortURL           string    `json:"short_url"`
	ExpiresAt          time.Time `json:"expires_at"`
	HasFixedExpiration bool      `json:"has_fixed_expiration"`
}

// GetURLResponse represents the response for getting URL metadata
type GetURLResponse struct {
	ShortCode          string    `json:"short_code"`
	OriginalURL        string    `json:"original_url"`
	Domain             string    `json:"domain"`
	CreatedAt          time.Time `json:"created_at"`
	ExpiresAt          time.Time `json:"expires_at"`
	HasFixedExpiration bool      `json:"has_fixed_expiration"`
	ClickCount         int64     `json:"click_count"`
	IsActive           bool      `json:"is_active"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}
