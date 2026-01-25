package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/application/queries"
)

// ListURLs handles GET /api/v1/urls
func (h *Handlers) ListURLs(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.Queries.ListURLs.Handle(r.Context(), queries.ListURLsQuery{})
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
