package apperror

import "errors"

var ErrNotFound = errors.New("Not Found")
var ErrConflict = errors.New("Conflict")
var ErrUnAuthorized = errors.New("UnAuthorized")
var ErrInvalidCredentials = errors.New("invalid credentials")

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func NewValidation(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}
