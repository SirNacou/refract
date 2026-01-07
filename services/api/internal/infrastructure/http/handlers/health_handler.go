package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/dto"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(pool *pgxpool.Pool, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		pool:   pool,
		logger: logger,
	}
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	ctx := r.Context()
	if err := h.pool.Ping(ctx); err != nil {
		h.logger.Error("Database health check failed", slog.String("error", err.Error()))
		writeJSON(w, http.StatusServiceUnavailable, dto.HealthResponse{
			Status: "unhealthy",
		})
		return
	}

	// Additional check: try to execute a simple query
	var result int
	if err := h.pool.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		h.logger.Error("Database query failed", slog.String("error", err.Error()))
		writeJSON(w, http.StatusServiceUnavailable, dto.HealthResponse{
			Status: "unhealthy",
		})
		return
	}

	writeJSON(w, http.StatusOK, dto.HealthResponse{
		Status: "ok",
	})
}

// Simple health check without database (for liveness probe)
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, dto.HealthResponse{
		Status: "ok",
	})
}

// Database readiness check
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*1000000000) // 2 seconds
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, dto.HealthResponse{
			Status: "not ready",
		})
		return
	}

	writeJSON(w, http.StatusOK, dto.HealthResponse{
		Status: "ready",
	})
}
