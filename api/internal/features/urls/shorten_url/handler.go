package shortenurl

import (
	"context"

	"github.com/SirNacou/refract/api/internal/infrastructure/auth"
)

type ShortenRequest struct {
	OriginalURL string `json:"original_url" format:"uri" required:"true"`
	Domain      string `json:"domain,omitempty"`
}

type ShortenResponse struct {
	Body *ShortenResponseBody
}

type ShortenResponseBody struct {
	ShortCode string `json:"short_code"`
	Domain    string `json:"domain"`
}

type Handler struct {
	cmd *CommandHandler
}

func NewHandler(cmd *CommandHandler) *Handler {
	return &Handler{
		cmd: cmd,
	}
}

func (h *Handler) Handle(ctx context.Context, req *struct {
	Body *ShortenRequest `json:"body" required:"true"`
}) (*ShortenResponse, error) {
	u, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	r, err := h.cmd.Handle(ctx, &Command{
		OriginalURL: req.Body.OriginalURL,
		Domain:      req.Body.Domain,
		UserID:      u,
	})
	if err != nil {
		return nil, err
	}
	// Implementation for shortening the URL and handling the domain
	return &ShortenResponse{
		Body: &ShortenResponseBody{
			ShortCode: r.ShortCode,
			Domain:    r.Domain,
		},
	}, nil
}
