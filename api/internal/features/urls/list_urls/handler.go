package listurls

import (
	"context"

	"github.com/SirNacou/refract/api/internal/infrastructure/auth"
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
	claims := auth.GetClaimsFromContext(ctx)
	userID, err := claims.GetUserID()
	if err != nil {
		return nil, err
	}

	res, err := h.query.Handle(ctx, &Query{userID})
	if err != nil {
		return nil, err
	}

	return &Response{Body: res}, nil
}
