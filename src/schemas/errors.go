package schemas

import (
	"fmt"
	"net/http"
)

// ErrorResponse represents a standard API error (RFC 7807).
// It implements the standard Go error interface.
type ErrorResponse struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"` // HTTP Status Code
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

// Error implements the error interface.
// This allows ErrorResponse to be returned as a standard Go error.
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Detail)
}

// NewErrorResponse creates a general ErrorResponse.
func NewErrorResponse(status int, title, detail, instance string) *ErrorResponse {
	return &ErrorResponse{
		Type:     fmt.Sprintf("https://connection-service.com/errors/%d", status),
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
}

// --- Helper Constructors for Common HTTP Errors ---

// NewBadRequestError creates a 400 Bad Request error.
func NewBadRequestError(detail, instance string) *ErrorResponse {
	return NewErrorResponse(http.StatusBadRequest, "Bad Request", detail, instance)
}

// NewNotFoundError creates a 404 Not Found error.
func NewNotFoundError(detail, instance string) *ErrorResponse {
	return NewErrorResponse(http.StatusNotFound, "Not Found", detail, instance)
}

// NewConflictError creates a 409 Conflict error.
func NewConflictError(detail, instance string) *ErrorResponse {
	return NewErrorResponse(http.StatusConflict, "Conflict", detail, instance)
}

// NewInternalError creates a 500 Internal Server Error.
// Note: Be careful not to expose sensitive technical details in production.
func NewInternalError(detail, instance string) *ErrorResponse {
	return NewErrorResponse(http.StatusInternalServerError, "Internal Server Error", detail, instance)
}

// NewBadGatewayError creates a 502 Bad Gateway error.
// Used when an upstream service (like users-service) fails.
func NewBadGatewayError(detail, instance string) *ErrorResponse {
	return NewErrorResponse(http.StatusBadGateway, "Bad Gateway", detail, instance)
}

// --- Domain-Specific Error Constructors ---

// SessionNotInProgressError creates a 409 Conflict error for session not in progress.
// This is a specialized error for the session workflow.
func SessionNotInProgressError(detail, instance string) *ErrorResponse {
	return &ErrorResponse{
		Type:     "https://connection-service.com/session-not-in-progress",
		Title:    "Session Not In Progress",
		Status:   http.StatusConflict,
		Detail:   detail,
		Instance: instance,
	}
}

// --- Backward Compatibility Helpers ---
// These can be removed once all code migrates to the New* constructors

// InternalServerError is a backward-compatible wrapper for NewInternalError.
// Deprecated: Use NewInternalError instead.
func InternalServerError(detail, instance string) *ErrorResponse {
	return NewInternalError(detail, instance)
}

// BadRequestError is a backward-compatible wrapper for NewBadRequestError.
// Deprecated: Use NewBadRequestError instead.
func BadRequestError(detail, instance string) *ErrorResponse {
	return NewBadRequestError(detail, instance)
}
