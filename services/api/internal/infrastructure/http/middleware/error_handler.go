package middleware

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/dto"
)

// ErrorHandler wraps domain errors into HTTP error responses
func ErrorHandler(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Store error in context if handler sets it
			next.ServeHTTP(w, r)
		})
	}
}

// HandleDomainError converts a domain error to an HTTP response
func HandleDomainError(w http.ResponseWriter, err error, logger *slog.Logger) {
	var domainErr *url.DomainError
	if errors.As(err, &domainErr) {
		logger.Warn("Domain error",
			slog.String("code", domainErr.Code),
			slog.String("message", domainErr.Message),
			slog.Int("status", domainErr.HTTPStatus()),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(domainErr.HTTPStatus())
		json.NewEncoder(w).Encode(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    domainErr.Code,
				Message: domainErr.Message,
			},
		})
		return
	}

	// Unknown error
	logger.Error("Unexpected error", slog.String("error", err.Error()))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred",
		},
	})
}
