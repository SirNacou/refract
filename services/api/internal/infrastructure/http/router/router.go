package router

import (
	"log/slog"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/handlers"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/middleware"
)

// Config holds router configuration
type Config struct {
	URLHandler    *handlers.URLHandler
	HealthHandler *handlers.HealthHandler
	Logger        *slog.Logger
	CORSOrigins   []string
}

// NewRouter creates and configures the HTTP router
func NewRouter(cfg Config) http.Handler {
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("GET /health", cfg.HealthHandler.Health)
	mux.HandleFunc("GET /health/live", cfg.HealthHandler.Liveness)
	mux.HandleFunc("GET /health/ready", cfg.HealthHandler.Readiness)

	// URL endpoints
	mux.HandleFunc("POST /api/urls", cfg.URLHandler.CreateURL)
	mux.HandleFunc("GET /api/urls/{shortCode}", cfg.URLHandler.GetURLMetadata)

	// Apply middleware stack (outermost to innermost)
	var handler http.Handler = mux

	// CORS (if origins configured)
	if len(cfg.CORSOrigins) > 0 {
		handler = middleware.CORS(cfg.CORSOrigins)(handler)
	}

	// Logger
	handler = middleware.Logger(cfg.Logger)(handler)

	// Recovery (outermost - catches all panics)
	handler = middleware.Recovery(cfg.Logger)(handler)

	return handler
}
