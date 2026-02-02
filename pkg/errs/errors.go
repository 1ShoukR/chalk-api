package errs

import (
	"errors"
	"net/http"
)

// Common application errors
var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrBadRequest        = errors.New("bad request")
	ErrInternalServer    = errors.New("internal server error")
	ErrConflict          = errors.New("resource conflict")
	ErrValidation        = errors.New("validation error")
	ErrDuplicateEntry    = errors.New("duplicate entry")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// AppError represents an application error with HTTP status
type AppError struct {
	Err        error
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Common error constructors
func NewNotFoundError(message string) *AppError {
	return NewAppError(ErrNotFound, message, http.StatusNotFound)
}

func NewBadRequestError(message string) *AppError {
	return NewAppError(ErrBadRequest, message, http.StatusBadRequest)
}

func NewUnauthorizedError(message string) *AppError {
	return NewAppError(ErrUnauthorized, message, http.StatusUnauthorized)
}

func NewForbiddenError(message string) *AppError {
	return NewAppError(ErrForbidden, message, http.StatusForbidden)
}

func NewInternalError(message string) *AppError {
	return NewAppError(ErrInternalServer, message, http.StatusInternalServerError)
}

func NewConflictError(message string) *AppError {
	return NewAppError(ErrConflict, message, http.StatusConflict)
}

func NewValidationError(message string) *AppError {
	return NewAppError(ErrValidation, message, http.StatusUnprocessableEntity)
}
