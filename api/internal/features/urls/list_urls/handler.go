package listurls

import "context"

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
	res, err := h.query.Handle(ctx, &Query{})
	if err != nil {
		return nil, err
	}

	return &Response{Body: res}, nil
}
