package errors

import (
	"log"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents the JSON structure for error responses
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   *AppError `json:"error"`
}

// ErrorHandler is a custom Echo error handler that formats errors consistently
func ErrorHandler(err error, c echo.Context) {
	var appErr *AppError

	// Check if it's already an AppError
	if IsAppError(err) {
		appErr = GetAppError(err)
	} else {
		// Convert standard errors to AppError
		appErr = NewInternalError("An unexpected error occurred", err)
	}

	// Log internal errors
	if appErr.Type == ErrorTypeInternal || appErr.Type == ErrorTypeExternal {
		log.Printf("Error: %v", appErr)
	}

	// Don't expose internal error details in production
	if appErr.Type == ErrorTypeInternal {
		appErr.Details = nil
		appErr.Cause = nil
	}

	// Send JSON error response
	if !c.Response().Committed {
		response := ErrorResponse{
			Success: false,
			Error:   appErr,
		}

		if err := c.JSON(appErr.HTTPStatus(), response); err != nil {
			log.Printf("Failed to send error response: %v", err)
		}
	}
}

// WrapError wraps any error into an AppError with context
func WrapError(err error, errorType ErrorType, message string) *AppError {
	if IsAppError(err) {
		return GetAppError(err)
	}

	return &AppError{
		Type:    errorType,
		Message: message,
		Cause:   err,
	}
}

// HandleValidationErrors converts validation errors to AppError
func HandleValidationErrors(validationErrors map[string]string) *AppError {
	return NewValidationError("Validation failed", validationErrors)
}
