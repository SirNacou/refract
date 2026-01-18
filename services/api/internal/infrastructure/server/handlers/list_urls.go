package handlers

import (
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/errors"
)

// ListURLs handles GET /api/v1/urls
func (h *Handlers) ListURLs(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement list URLs query handler
	errors.WriteInternalError(w, r, "Not implemented yet")
}
