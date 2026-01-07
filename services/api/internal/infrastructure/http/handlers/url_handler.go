package handlers

import (
	"log/slog"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/application/commands"
	"github.com/SirNacou/refract/services/api/internal/application/queries"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/dto"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/middleware"
)

// URLHandler handles URL-related HTTP requests
type URLHandler struct {
	createHandler *commands.CreateURLHandler
	getHandler    *queries.GetURLHandler
	logger        *slog.Logger
}

// NewURLHandler creates a new URL handler
func NewURLHandler(
	createHandler *commands.CreateURLHandler,
	getHandler *queries.GetURLHandler,
	logger *slog.Logger,
) *URLHandler {
	return &URLHandler{
		createHandler: createHandler,
		getHandler:    getHandler,
		logger:        logger,
	}
}

// CreateURL handles POST /api/urls
func (h *URLHandler) CreateURL(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateURLRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "Invalid JSON request body",
			},
		})
		return
	}

	cmd := commands.CreateURLCommand{
		OriginalURL: req.OriginalURL,
		Domain:      req.Domain,
		ExpiresAt:   req.ExpiresAt,
	}

	result, err := h.createHandler.Handle(r.Context(), cmd)
	if err != nil {
		middleware.HandleDomainError(w, err, h.logger)
		return
	}

	response := dto.CreateURLResponse{
		ShortCode:          result.ShortCode,
		ShortURL:           result.ShortURL,
		ExpiresAt:          result.ExpiresAt,
		HasFixedExpiration: result.HasFixedExpiration,
	}

	writeJSON(w, http.StatusCreated, response)
}

// GetURLMetadata handles GET /api/urls/{shortCode}
func (h *URLHandler) GetURLMetadata(w http.ResponseWriter, r *http.Request) {
	// Extract short code from path
	shortCode := r.PathValue("shortCode")
	if shortCode == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "MISSING_SHORT_CODE",
				Message: "Short code is required",
			},
		})
		return
	}

	query := queries.GetURLQuery{
		ShortCode: shortCode,
	}

	result, err := h.getHandler.Handle(r.Context(), query)
	if err != nil {
		middleware.HandleDomainError(w, err, h.logger)
		return
	}

	response := dto.GetURLResponse{
		ShortCode:          result.ShortCode,
		OriginalURL:        result.OriginalURL,
		Domain:             result.Domain,
		CreatedAt:          result.CreatedAt,
		ExpiresAt:          result.ExpiresAt,
		HasFixedExpiration: result.HasFixedExpiration,
		ClickCount:         result.ClickCount,
		IsActive:           result.IsActive,
	}

	writeJSON(w, http.StatusOK, response)
}
