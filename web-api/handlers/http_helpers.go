package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse is the canonical error payload shape for all non-2xx responses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// WriteJSON writes a JSON response with the provided status code and payload.
// It sets the Content-Type to application/json and disables HTML escaping.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(v); err != nil {
		// We can't change headers at this point; just log the failure.
		log.Printf("WriteJSON encode error: %v", err)
	}
}

// WriteError writes a standardized JSON error payload with the given status code.
// Use this for all error responses to keep a consistent API shape.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, ErrorResponse{Error: msg})
}

// WriteErrorWithCode writes a standardized JSON error payload with an error code.
// The code can be used by clients for programmatic error handling.
func WriteErrorWithCode(w http.ResponseWriter, status int, code, msg string) {
	WriteJSON(w, status, ErrorResponse{Error: msg, Code: code})
}

// WriteValidationError writes a 422 Unprocessable Entity response with field-level details.
// 'details' can be a map[string]string or any structure describing validation failures.
func WriteValidationError(w http.ResponseWriter, details any) {
	WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
		Error:   "validation failed",
		Code:    "validation_error",
		Details: details,
	})
}
