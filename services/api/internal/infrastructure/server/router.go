package server

import (
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/application"
	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/handlers"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	authMw *middleware.AuthMiddleware,
	rateLimiter *middleware.RateLimiter,
	logging *middleware.LoggingMiddleware,
	securityCfg *config.SecurityConfig,
	app *application.Application,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.NewCORSHandler(securityCfg))
	r.Use(chimw.RequestID)
	r.Use(logging.Handler())
	r.Use(chimw.Recoverer)

	r.Get("/health", healthHandler)

	// Create handlers instance
	h := handlers.NewHandlers(app)

	r.Route("/api/v1", func(apiRouter chi.Router) {
		apiRouter.Use(authMw.RequireAuthorization())
		apiRouter.Use(rateLimiter.RateLimitPerUser())

		// Register all routes
		h.RegisterRoutes(apiRouter)
	})

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
