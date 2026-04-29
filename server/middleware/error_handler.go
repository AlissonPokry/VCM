package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
)

var errorCodes = map[int]string{
	http.StatusBadRequest:          "BAD_REQUEST",
	http.StatusUnauthorized:        "UNAUTHORIZED",
	http.StatusNotFound:            "NOT_FOUND",
	http.StatusConflict:            "CONFLICT",
	http.StatusRequestEntityTooLarge: "PAYLOAD_TOO_LARGE",
	http.StatusUnprocessableEntity: "VALIDATION_ERROR",
	http.StatusInternalServerError: "INTERNAL_SERVER_ERROR",
}

// Meta contains response metadata shared by list endpoints.
type Meta struct {
	Total int `json:"total"`
}

// APIResponse is the success response envelope used by all handlers.
type APIResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
	Meta    any  `json:"meta,omitempty"`
}

// APIError is the error response envelope used by all handlers.
type APIError struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// AppError carries an HTTP status and stable API error code.
type AppError struct {
	Message string
	Status  int
	Code    string
}

// Error returns the API-safe error message.
func (e AppError) Error() string {
	return e.Message
}

// NewAppError creates an application error with a stable code.
func NewAppError(message string, status int, code string) AppError {
	if code == "" {
		code = errorCodes[status]
	}
	if code == "" {
		code = "INTERNAL_SERVER_ERROR"
	}
	return AppError{Message: message, Status: status, Code: code}
}

// Respond writes a JSON success response envelope.
func Respond(w http.ResponseWriter, status int, data any, meta any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse{Success: true, Data: data, Meta: meta})
}

// RespondError writes a JSON error response envelope.
func RespondError(w http.ResponseWriter, status int, message string, code string) {
	if code == "" {
		code = errorCodes[status]
	}
	if code == "" {
		code = "INTERNAL_SERVER_ERROR"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIError{Error: true, Message: message, Code: code})
}

// HandleError maps an error to the API error envelope.
func HandleError(w http.ResponseWriter, err error) {
	var appErr AppError
	if errors.As(err, &appErr) {
		RespondError(w, appErr.Status, appErr.Message, appErr.Code)
		return
	}
	RespondError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_SERVER_ERROR")
}
