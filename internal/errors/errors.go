package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "VALIDATION_ERROR"
	ErrorTypeAuthentication ErrorType = "AUTHENTICATION_ERROR"
	ErrorTypeAuthorization  ErrorType = "AUTHORIZATION_ERROR"
	ErrorTypeNotFound       ErrorType = "NOT_FOUND"
	ErrorTypeConflict       ErrorType = "CONFLICT"
	ErrorTypeInternal       ErrorType = "INTERNAL_ERROR"
	ErrorTypeExternal       ErrorType = "EXTERNAL_ERROR"
	ErrorTypeRateLimit      ErrorType = "RATE_LIMIT_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Type    ErrorType   `json:"type"`
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
	Cause   error       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause of the error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *AppError) HTTPStatus() int {
	switch e.Type {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeExternal:
		return http.StatusBadGateway
	case ErrorTypeInternal:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details interface{}) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    "VALIDATION_FAILED",
		Details: details,
	}
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeAuthentication,
		Message: message,
		Code:    "AUTHENTICATION_FAILED",
	}
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeAuthorization,
		Message: message,
		Code:    "AUTHORIZATION_FAILED",
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Code:    "RESOURCE_NOT_FOUND",
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: message,
		Code:    "RESOURCE_CONFLICT",
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Code:    "INTERNAL_ERROR",
		Cause:   cause,
	}
}

// NewExternalError creates a new external service error
func NewExternalError(message string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeExternal,
		Message: message,
		Code:    "EXTERNAL_SERVICE_ERROR",
		Cause:   cause,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeRateLimit,
		Message: message,
		Code:    "RATE_LIMIT_EXCEEDED",
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError safely extracts an AppError from an error
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}
