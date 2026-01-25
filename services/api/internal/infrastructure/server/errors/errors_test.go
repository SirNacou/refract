package errors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
)

// Helper to create request with optional request ID in context
func newRequestWithID(requestID string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	if requestID != "" {
		ctx := context.WithValue(r.Context(), middleware.RequestIDKey, requestID)
		r = r.WithContext(ctx)
	}
	return r
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		errCode       string
		message       string
		requestID     string
		wantRequestID bool
	}{
		{
			name:          "basic error without request ID",
			statusCode:    http.StatusBadRequest,
			errCode:       ErrCodeInvalidRequest,
			message:       "Invalid input",
			requestID:     "",
			wantRequestID: false,
		},
		{
			name:          "error with request ID",
			statusCode:    http.StatusNotFound,
			errCode:       ErrCodeNotFound,
			message:       "Resource not found",
			requestID:     "req-123-456",
			wantRequestID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := newRequestWithID(tt.requestID)

			WriteError(w, r, tt.statusCode, tt.errCode, tt.message)

			// Check status code
			if w.Code != tt.statusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.statusCode)
			}

			// Check Content-Type header
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
			}

			// Parse response body
			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			// Check error code
			if resp.Error != tt.errCode {
				t.Errorf("error code = %q, want %q", resp.Error, tt.errCode)
			}

			// Check message
			if resp.Message != tt.message {
				t.Errorf("message = %q, want %q", resp.Message, tt.message)
			}

			// Check request ID
			if tt.wantRequestID {
				if resp.RequestID != tt.requestID {
					t.Errorf("request_id = %q, want %q", resp.RequestID, tt.requestID)
				}
			} else {
				if resp.RequestID != "" {
					t.Errorf("request_id = %q, want empty", resp.RequestID)
				}
			}

			// Details should be nil for WriteError
			if resp.Details != nil {
				t.Errorf("details = %v, want nil", resp.Details)
			}
		})
	}
}

func TestWriteErrorWithDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r := newRequestWithID("req-abc")

	details := map[string]any{
		"field":  "email",
		"reason": "invalid format",
	}

	WriteErrorWithDetails(w, r, http.StatusBadRequest, ErrCodeInvalidRequest, "Validation failed", details)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	// Parse response body
	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check details are included
	if resp.Details == nil {
		t.Fatal("details = nil, want non-nil")
	}

	if resp.Details["field"] != "email" {
		t.Errorf("details[field] = %v, want %q", resp.Details["field"], "email")
	}

	if resp.Details["reason"] != "invalid format" {
		t.Errorf("details[reason] = %v, want %q", resp.Details["reason"], "invalid format")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	tests := []struct {
		name        string
		fn          func(http.ResponseWriter, *http.Request, string)
		wantStatus  int
		wantErrCode string
	}{
		{
			name:        "WriteBadRequest",
			fn:          WriteBadRequest,
			wantStatus:  http.StatusBadRequest,
			wantErrCode: ErrCodeInvalidRequest,
		},
		{
			name:        "WriteUnauthorized",
			fn:          WriteUnauthorized,
			wantStatus:  http.StatusUnauthorized,
			wantErrCode: ErrCodeUnauthorized,
		},
		{
			name:        "WriteForbidden",
			fn:          WriteForbidden,
			wantStatus:  http.StatusForbidden,
			wantErrCode: ErrCodeForbidden,
		},
		{
			name:        "WriteNotFound",
			fn:          WriteNotFound,
			wantStatus:  http.StatusNotFound,
			wantErrCode: ErrCodeNotFound,
		},
		{
			name:        "WriteRateLimitExceeded",
			fn:          WriteRateLimitExceeded,
			wantStatus:  http.StatusTooManyRequests,
			wantErrCode: ErrCodeRateLimitExceeded,
		},
		{
			name:        "WriteInternalError",
			fn:          WriteInternalError,
			wantStatus:  http.StatusInternalServerError,
			wantErrCode: ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := newRequestWithID("")

			tt.fn(w, r, "test message")

			// Check status code
			if w.Code != tt.wantStatus {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatus)
			}

			// Parse response body
			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			// Check error code
			if resp.Error != tt.wantErrCode {
				t.Errorf("error code = %q, want %q", resp.Error, tt.wantErrCode)
			}

			// Check message
			if resp.Message != "test message" {
				t.Errorf("message = %q, want %q", resp.Message, "test message")
			}
		})
	}
}

func TestErrorResponseJSONStructure(t *testing.T) {
	// Test that JSON output matches OpenAPI spec structure
	w := httptest.NewRecorder()
	r := newRequestWithID("req-xyz")

	WriteError(w, r, http.StatusNotFound, ErrCodeNotFound, "URL not found")

	// Parse as raw JSON to check field names
	var raw map[string]any
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check required fields exist with correct names (per OpenAPI spec)
	if _, ok := raw["error"]; !ok {
		t.Error("missing 'error' field in JSON response")
	}

	if _, ok := raw["message"]; !ok {
		t.Error("missing 'message' field in JSON response")
	}

	if _, ok := raw["request_id"]; !ok {
		t.Error("missing 'request_id' field in JSON response")
	}

	// Check 'details' is omitted when nil (omitempty)
	if _, ok := raw["details"]; ok {
		t.Error("'details' field should be omitted when nil")
	}
}
