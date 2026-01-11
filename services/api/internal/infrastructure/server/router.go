package server

import (
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	authMw *middleware.AuthMiddleware,
	rateLimiter *middleware.RateLimiter,
	logging *middleware.LoggingMiddleware,
	securityCfg *config.SecurityConfig,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.NewCORSHandler(securityCfg))
	r.Use(chimw.RequestID)
	r.Use(logging.Handler())
	r.Use(chimw.Recoverer)

	r.Get("/health", healthHandler)

	r.Route("/api/v1", func(apiRouter chi.Router) {
		// Define your API routes here
		apiRouter.Use(authMw.RequireAuthentication())
		apiRouter.Use(rateLimiter.RateLimitPerUser())

		apiRouter.Get("/urls", listURLsHandler)
		apiRouter.Post("/urls", createURLHandler)

		apiRouter.Get("/analytics/{id}", getAnalyticsHandler)
	})

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func listURLsHandler(w http.ResponseWriter, r *http.Request) {
	// Handler logic for listing URLs
}

func createURLHandler(w http.ResponseWriter, r *http.Request) {
	// Handler logic for creating a new URL
}

func getAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	// Handler logic for getting analytics by ID
}
