package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/dto"
)

// Recovery recovers from panics and logs them
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()

					logger.Error("Panic recovered",
						slog.Any("error", err),
						slog.String("stack", string(stack)),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					response := dto.ErrorResponse{
						Error: dto.ErrorDetail{
							Code:    "INTERNAL_ERROR",
							Message: fmt.Sprintf("Internal server error: %v", err),
						},
					}

					// Best effort to write response
					_ = json.NewEncoder(w).Encode(response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
