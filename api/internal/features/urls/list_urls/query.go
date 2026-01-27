package listurls

import "context"

type Query struct{
	userID string
}

type QueryResponse struct{}

type QueryHandler struct{}

func NewQueryHandler() *QueryHandler {
	return &QueryHandler{}
}

func (h *QueryHandler) Handle(ctx context.Context, req *Query) (*QueryResponse, error) {

	return &QueryResponse{}, nil
}
