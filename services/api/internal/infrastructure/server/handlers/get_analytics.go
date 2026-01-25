package handlers

import (
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/errors"
)

// GetAnalytics handles GET /api/v1/analytics/{id}
func (h *Handlers) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement analytics query handler
	errors.WriteInternalError(w, r, "Not implemented yet")
}
