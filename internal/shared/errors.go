package shared

import (
	"errors"
	"net/http"
)

// Error codes
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeConflict      = "CONFLICT"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeBadRequest    = "BAD_REQUEST"
	ErrCodeUnprocessable = "UNPROCESSABLE_ENTITY"

	// Repository error codes
	ErrCodeUserNotFound  = "USER_NOT_FOUND"
	ErrCodeUserExists    = "USER_EXISTS"
	ErrCodeTokenNotFound = "TOKEN_NOT_FOUND"
	ErrCodeTokenExpired  = "TOKEN_EXPIRED"
	ErrCodeTokenUsed     = "TOKEN_USED"
	ErrCodeTokenInvalid  = "TOKEN_INVALID"

	// Profile error codes
	ErrCodeProfileNotFound = "PROFILE_NOT_FOUND"
)

// AppError represents an application error with status code and error code
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Details    map[string]interface{}
	Err        error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	e.Details[key] = value
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToResponse converts AppError to ErrorResponse
func (e *AppError) ToResponse() ErrorResponse {
	details := e.Details
	if details == nil {
		details = make(map[string]interface{})
	}

	return ErrorResponse{
		Error:   http.StatusText(e.StatusCode),
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// Predefined errors
var (
	ErrValidation    = NewAppError(ErrCodeValidation, "Validation error", http.StatusBadRequest)
	ErrNotFound      = NewAppError(ErrCodeNotFound, "Resource not found", http.StatusNotFound)
	ErrUnauthorized  = NewAppError(ErrCodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden     = NewAppError(ErrCodeForbidden, "Forbidden", http.StatusForbidden)
	ErrConflict      = NewAppError(ErrCodeConflict, "Resource conflict", http.StatusConflict)
	ErrInternal      = NewAppError(ErrCodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrBadRequest    = NewAppError(ErrCodeBadRequest, "Bad request", http.StatusBadRequest)
	ErrUnprocessable = NewAppError(ErrCodeUnprocessable, "Unprocessable entity", http.StatusUnprocessableEntity)

	// Repository errors
	ErrUserNotFound  = NewAppError(ErrCodeUserNotFound, "User not found", http.StatusNotFound)
	ErrUserExists    = NewAppError(ErrCodeUserExists, "User already exists", http.StatusConflict)
	ErrTokenNotFound = NewAppError(ErrCodeTokenNotFound, "Token not found", http.StatusNotFound)
	ErrTokenExpired  = NewAppError(ErrCodeTokenExpired, "Token has expired", http.StatusUnauthorized)
	ErrTokenUsed     = NewAppError(ErrCodeTokenUsed, "Token has already been used", http.StatusUnauthorized)
	ErrTokenInvalid  = NewAppError(ErrCodeTokenInvalid, "Invalid token", http.StatusUnauthorized)

	// Profile errors
	ErrProfileNotFound = NewAppError(ErrCodeProfileNotFound, "Profile not found", http.StatusNotFound)
)

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// ToAppError converts an error to AppError
func ToAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	// Try to unwrap
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Return internal error as fallback
	return ErrInternal.WithError(err)
}
