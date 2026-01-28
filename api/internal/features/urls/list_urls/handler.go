package listurls

import (
	"context"

	"github.com/SirNacou/refract/api/internal/infrastructure/auth"
	"github.com/danielgtaylor/huma/v2"
)

type Request struct{}

type Response struct {
	Body *QueryResponse
}

type Handler struct {
	query *QueryHandler
}

func NewHandler(query *QueryHandler) *Handler {
	return &Handler{query: query}
}

func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	claims, err := auth.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, huma.Error401Unauthorized("Unauthorized", err)
	}

	userID, err := claims.GetUserID()
	if err != nil {
		return nil, huma.Error401Unauthorized("Unauthorized", err)
	}

	res, err := h.query.Handle(ctx, &Query{userID})
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to query URLs", err)
	}

	return &Response{Body: res}, nil
}
