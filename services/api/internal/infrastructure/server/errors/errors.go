package errors

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type ErrorResponse struct {
	Error     string         `json:"error"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
}

const (
	ErrCodeInvalidRequest    = "INVALID_REQUEST"
	ErrCodeInvalidURL        = "INVALID_URL"
	ErrCodeAliasTaken        = "ALIAS_TAKEN"
	ErrCodeMaliciousURL      = "MALICIOUS_URL"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeForbidden         = "FORBIDDEN"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternalError     = "INTERNAL_ERROR"
)

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	requestID := middleware.GetReqID(r.Context())
	errorResponse := ErrorResponse{
		Error:     errCode,
		Message:   message,
		RequestID: requestID,
	}
	_ = json.NewEncoder(w).Encode(errorResponse)
}

func WriteErrorWithDetails(w http.ResponseWriter, r *http.Request, statusCode int, errCode, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	requestID := middleware.GetReqID(r.Context())
	errorResponse := ErrorResponse{
		Error:     errCode,
		Message:   message,
		Details:   details,
		RequestID: requestID,
	}
	_ = json.NewEncoder(w).Encode(errorResponse)
}
func WriteBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusBadRequest, ErrCodeInvalidRequest, message)
}

func WriteUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

func WriteForbidden(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusForbidden, ErrCodeForbidden, message)
}

func WriteNotFound(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusNotFound, ErrCodeNotFound, message)
}

func WriteRateLimitExceeded(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusTooManyRequests, ErrCodeRateLimitExceeded, message)
}

func WriteInternalError(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusInternalServerError, ErrCodeInternalError, message)
}
