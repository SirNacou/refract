package listurls

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SirNacou/refract/api/internal/domain"
)

type Query struct {
	userID string
}

type QueryResponse struct {
	URLs []URL `json:"urls" default:"[]"`
}

type URL struct {
	ID          string        `json:"id"`
	OriginalURL string        `json:"original_url"`
	ShortURL    string        `json:"short_url"`
	Title       string        `json:"title"`
	Notes       string        `json:"notes"`
	UserID      string        `json:"user_id"`
	ExpiresAt   *time.Time    `json:"expires_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Status      domain.Status `json:"status"`
}

type QueryHandler struct {
	repo           domain.URLRepository
	defaultBaseURL string
}

func NewQueryHandler(repo domain.URLRepository, defaultBaseURL string) *QueryHandler {
	return &QueryHandler{
		repo:           repo,
		defaultBaseURL: defaultBaseURL,
	}
}

func (h *QueryHandler) Handle(ctx context.Context, req *Query) (*QueryResponse, error) {
	urls, err := h.repo.ListByUser(ctx, req.userID)
	if err != nil {
		return nil, err
	}

	converted := make([]URL, len(urls))
	for i, u := range urls {
		sURL := strings.Join([]string{h.defaultBaseURL, u.ShortCode.String()}, "/")
		converted[i] = URL{
			ID:          fmt.Sprint(u.ID.Int64()),
			OriginalURL: u.OriginalURL,
			ShortURL:    sURL,
			Title:       u.Title,
			Notes:       u.Notes,
			UserID:      u.UserID,
			ExpiresAt:   u.ExpiresAt,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
			Status:      u.Status,
		}
	}

	return &QueryResponse{
		URLs: converted,
	}, nil
}
