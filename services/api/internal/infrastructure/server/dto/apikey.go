package dto

import "time"

// Requests
type GenerateAPIKeyRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// Responses
type GenerateAPIKeyResponse struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"` // Full key, shown only once
	KeyPrefix string    `json:"key_prefix"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type APIKeyResponse struct {
	ID         int64      `json:"id"`
	KeyPrefix  string     `json:"key_prefix"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	UsageCount int64      `json:"usage_count"`
}

type APIKeyListResponse struct {
	APIKeys []APIKeyResponse `json:"api_keys"`
}
