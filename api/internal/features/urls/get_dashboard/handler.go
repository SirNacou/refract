package getdashboard

import (
	"context"

	"github.com/SirNacou/refract/api/internal/infrastructure/auth"
	"github.com/danielgtaylor/huma/v2"
)

type DashboardRequest struct {
}

type DashboardResponse struct {
	Body *QueryResult
}

type Handler struct {
	query *QueryHandler
}

func NewHandler(query *QueryHandler) *Handler {
	return &Handler{query: query}
}

func (h *Handler) Handle(ctx context.Context, q *DashboardRequest) (*DashboardResponse, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, huma.Error401Unauthorized("Failed to get user id", err)
	}

	res, err := h.query.Handle(ctx, &Query{
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	return &DashboardResponse{
		Body: res,
	}, nil
}
