package response

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// SuccessEnvelope wraps a successful response.
type SuccessEnvelope struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// ErrorEnvelope wraps an error response.
type ErrorEnvelope struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
	Details any    `json:"details"`
}

// Success writes a standardized success response.
func Success(w http.ResponseWriter, status int, data any, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if message == "" {
		message = http.StatusText(status)
	}
	_ = json.NewEncoder(w).Encode(SuccessEnvelope{
		Status: "success", Code: status, Message: message, Data: data,
	})
}

// Error writes a standardized error response with a machine-readable code.
func Error(w http.ResponseWriter, status int, message string, machineCode string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{
		Status: "error", Code: status, Message: message, Error: machineCode, Details: nil,
	})
}

// ErrorCode maps HTTP status to machine-readable error code.
func ErrorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusRequestEntityTooLarge:
		return "REQUEST_TOO_LARGE"
	case http.StatusTooManyRequests:
		return "RATE_LIMITED"
	case http.StatusBadGateway:
		return "UPSTREAM_ERROR"
	case http.StatusGatewayTimeout:
		return "TIMEOUT"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// ErrorSimple writes an error with machine code derived from status.
func ErrorSimple(w http.ResponseWriter, status int, message string) {
	Error(w, status, message, ErrorCode(status))
}

// ErrorValidation writes a 400 validation error with specific field details.
func ErrorValidation(w http.ResponseWriter, message string, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{
		Status:  "error",
		Code:    http.StatusBadRequest,
		Message: message,
		Error:   "VALIDATION_ERROR",
		Details: map[string]string{"field": field},
	})
}

// JSON is deprecated — use Success instead. Kept for internal backward compat.
func JSON(w http.ResponseWriter, status int, data any) {
	Success(w, status, data, "")
}

// NoContent writes a 204 No Content response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// DecodeJSON decodes the JSON request body into the provided destination.
func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("request body is empty")
	}
	return json.Unmarshal(body, dst)
}
