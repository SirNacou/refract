package dto

import "time"

// Requests
type CreateURLRequest struct {
	DestinationURL string     `json:"destination_url" validate:"required,url,max=2048"`
	ShortCode      *string    `json:"short_code,omitempty" validate:"omitempty,min=3,max=50"`
	Title          string     `json:"title" validate:"required,max=200"`
	Notes          *string    `json:"notes,omitempty" validate:"omitempty,max=1000"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type UpdateURLRequest struct {
	DestinationURL *string    `json:"destination_url,omitempty" validate:"omitempty,url,max=2048"`
	Title          *string    `json:"title,omitempty" validate:"omitempty,max=200"`
	Notes          *string    `json:"notes,omitempty" validate:"omitempty,max=1000"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// Responses
type CreateURLResponse struct {
	ShortCode string `json:"short_code"`
}

type URLResponse struct {
	ID             int64      `json:"id"`
	ShortCode      string     `json:"short_code"`
	ShortURL       string     `json:"short_url"`
	DestinationURL string     `json:"destination_url"`
	Title          *string    `json:"title,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	Status         string     `json:"status"`
	TotalClicks    int64      `json:"total_clicks"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	LastClickedAt  *time.Time `json:"last_clicked_at,omitempty"`
}

type URLListResponse struct {
	URLs       []URLResponse  `json:"urls"`
	Pagination PaginationMeta `json:"pagination"`
}
