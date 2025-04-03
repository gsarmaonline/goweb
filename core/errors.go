package core

import "fmt"

// ErrInvalidField represents a validation error for a specific field
type ErrInvalidField struct {
	Field   string
	Message string
}

func (e ErrInvalidField) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ErrDeleteForbidden represents an error when deletion is not allowed
type ErrDeleteForbidden struct {
	Message string
}

func (e ErrDeleteForbidden) Error() string {
	return e.Message
}
