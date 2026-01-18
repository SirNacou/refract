package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/application/commands"
	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/dto"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/errors"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/validator"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
)

// CreateURL handles POST /api/v1/urls
func (h *Handlers) CreateURL(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req dto.CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate request
	if err := validator.Validate.Struct(req); err != nil {
		errors.WriteBadRequest(w, r, "Validation failed: "+err.Error())
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID := authorization.UserID(r.Context())
	if userID == "" {
		errors.WriteUnauthorized(w, r, "User not authenticated")
		return
	}

	// Map DTO to command
	cmd := commands.CreateURLCommand{
		CustomAlias:    req.CustomAlias,
		DestinationURL: req.DestinationURL,
		Title:          getStringValue(req.Title),
		Notes:          getStringValue(req.Notes),
		ExpiresAt:      req.ExpiresAt,
		CreatorUserID:  userID,
	}

	// Execute command
	result, err := h.app.Commands.CreateURL.Handle(cmd)
	if err != nil {
		handleCommandError(w, r, err)
		return
	}

	// Return response
	response := dto.CreateURLResponse{
		ShortCode: result.ShortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleCommandError maps domain errors to HTTP responses
func handleCommandError(w http.ResponseWriter, r *http.Request, err error) {
	switch err {
	case url.ErrAliasAlreadyTaken:
		errors.WriteError(w, r, http.StatusConflict, errors.ErrCodeAliasTaken, "Alias already taken")
	case url.ErrInvalidURL:
		errors.WriteError(w, r, http.StatusBadRequest, errors.ErrCodeInvalidURL, "Invalid URL")
	case url.ErrInvalidExpiry:
		errors.WriteBadRequest(w, r, "Invalid expiry date")
	case url.ErrInvalidShortCode:
		errors.WriteBadRequest(w, r, "Invalid short code")
	case url.ErrMaliciousURL:
		errors.WriteError(w, r, http.StatusBadRequest, errors.ErrCodeMaliciousURL, "Malicious URL detected")
	default:
		errors.WriteInternalError(w, r, "An internal error occurred")
	}
}

// getStringValue safely extracts string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
