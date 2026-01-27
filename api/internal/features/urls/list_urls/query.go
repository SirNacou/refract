package listurls

import (
	"context"

	"github.com/SirNacou/refract/api/internal/domain"
)

type Query struct {
	userID string
}

type QueryResponse struct {
	URLs []domain.URL `json:"urls"`
}

type QueryHandler struct {
	repo domain.URLRepository
}

func NewQueryHandler(repo domain.URLRepository) *QueryHandler {
	return &QueryHandler{
		repo: repo,
	}
}

func (h *QueryHandler) Handle(ctx context.Context, req *Query) (*QueryResponse, error) {
	urls, err := h.repo.ListByUser(ctx, req.userID)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{
		URLs: urls,
	}, nil
}
