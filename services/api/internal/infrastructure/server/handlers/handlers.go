package handlers

import (
	"github.com/SirNacou/refract/services/api/internal/application"
	"github.com/go-chi/chi/v5"
)

// Handlers holds references to all HTTP handlers
type Handlers struct {
	app *application.Application
}

// NewHandlers creates a new Handlers instance
func NewHandlers(app *application.Application) *Handlers {
	return &Handlers{app: app}
}

// RegisterRoutes registers all HTTP routes with the router
func (h *Handlers) RegisterRoutes(r chi.Router) {
	// URL management routes
	r.Post("/urls", h.CreateURL)
	r.Get("/urls", h.ListURLs)

	// Analytics routes
	r.Get("/analytics/{id}", h.GetAnalytics)
}
