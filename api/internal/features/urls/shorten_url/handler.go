package shortenurl

import (
	"context"

	"github.com/SirNacou/refract/api/internal/infrastructure/auth"
)

type ShortenRequest struct {
	Title       string `json:"title" maxLength:"255" required:"true"`
	OriginalURL string `json:"original_url" format:"uri" required:"true"`
	CustomAlias *string `json:"custom_alias" maxLength:"20" required:"false"`
}

type ShortenResponse struct {
	Body *ShortenResponseBody
}

type ShortenResponseBody struct {
	ShortURL string `json:"short_url"`
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
		UserID:      u,
		Title:       req.Body.Title,
		CustomAlias: req.Body.CustomAlias,
	})
	if err != nil {
		return nil, err
	}
	// Implementation for shortening the URL and handling the domain
	return &ShortenResponse{
		Body: &ShortenResponseBody{
			ShortURL: r.ShortURL,
		},
	}, nil
}
