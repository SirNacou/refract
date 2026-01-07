package url

import "fmt"

// ErrorType represents the category of domain error
type ErrorType int

const (
	ErrorTypeNotFound ErrorType = iota
	ErrorTypeValidation
	ErrorTypeConflict
	ErrorTypeInternal
)

// DomainError represents a business logic error
type DomainError struct {
	Type    ErrorType
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap implements error unwrapping
func (e *DomainError) Unwrap() error {
	return e.Err
}

// HTTPStatus maps error types to HTTP status codes
func (e *DomainError) HTTPStatus() int {
	switch e.Type {
	case ErrorTypeNotFound:
		return 404
	case ErrorTypeValidation:
		return 400
	case ErrorTypeConflict:
		return 409
	case ErrorTypeInternal:
		return 500
	default:
		return 500
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeNotFound,
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Code:    code,
		Message: message,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeConflict,
		Code:    code,
		Message: message,
	}
}

// NewInternalError creates an internal error
func NewInternalError(code, message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternal,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrap wraps an existing error as a domain error
func WrapError(errType ErrorType, code, message string, err error) *DomainError {
	return &DomainError{
		Type:    errType,
		Code:    code,
		Message: message,
		Err:     err,
	}
}
