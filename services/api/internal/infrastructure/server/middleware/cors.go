package middleware

import (
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/go-chi/cors"
)

func NewCORSHandler(cfg *config.SecurityConfig) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedHeaders:   cfg.CORSAllowedHeaders,
		AllowedMethods:   cfg.CORSAllowedMethods,
		AllowCredentials: true,
		MaxAge:           300,
	})
}
