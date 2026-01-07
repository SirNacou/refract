package handlers

import (
	"encoding/json"
	"io"
	"net/http"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but can't change response at this point
		_ = err
	}
}

// decodeJSON decodes JSON request body
func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()

	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(nil, r.Body, 1048576)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return err
	}

	// Ensure there's no additional data after the JSON object
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return err
	}

	return nil
}
