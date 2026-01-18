package url

import (
	"net/url"
	"time"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/validator"
)

type Status string

const (
	Active   Status = "active"
	Expired  Status = "expired"
	Disabled Status = "disabled"
	Deleted  Status = "deleted"
)

type URL struct {
	ID             uint64     `json:"id"`
	CustomAlias    ShortCode  `json:"custom_alias"`
	DestinationURL string     `json:"destination_url"`
	Title          string     `json:"title"`
	Notes          string     `json:"notes"`
	Status         Status     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatorUserID  string     `json:"creator_user_id"`
	TotalClicks    uint64     `json:"total_clicks"`
	LastClickedAt  *time.Time `json:"last_clicked_at"`
}

type CreateURLRequest struct {
	ID             uint64 `validate:"required"`
	CustomAlias    *ShortCode
	DestinationURL string `validate:"url"`
	Title          string `validate:"required"`
	Notes          string `validate:"max=255"`
	ExpiresAt      *time.Time
	CreatorUserID  string `validate:"required"`
}

func NewURL(req CreateURLRequest) (*URL, error) {
	err := validator.Validate.Struct(req)
	if err != nil {
		return nil, err
	}

	var customAlias ShortCode
	if req.CustomAlias != nil {
		customAlias = *req.CustomAlias
	} else {
		customAlias = *NewShortCode(req.ID)
	}

	if !req.ExpiresAt.After(time.Now().Add(time.Minute)) {
		return nil, ErrInvalidExpiry
	}

	e := &URL{
		ID:             req.ID,
		CustomAlias:    customAlias,
		DestinationURL: req.DestinationURL,
		Title:          req.Title,
		Notes:          req.Notes,
		Status:         Active,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ExpiresAt:      req.ExpiresAt,
		CreatorUserID:  req.CreatorUserID,
		TotalClicks:    0,
		LastClickedAt:  nil,
	}

	err = e.ValidateDestinationURL()
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (u *URL) ValidateDestinationURL() error {
	if u.DestinationURL == "" {
		return ErrInvalidURL
	}
	parsed, err := url.ParseRequestURI(u.DestinationURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidURL
	}
	return nil
}
