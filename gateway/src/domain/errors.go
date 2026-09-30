package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrCircuitOpen        = fmt.Errorf("circuit open: %w", ErrServiceUnavailable)
)

type Error struct {
	Kind    error
	Message string
	Cause   error
}

func NewError(kind error, format string, args ...any) *Error {
	return &Error{Kind: kind, Message: fmt.Sprintf(format, args...)}
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Kind
}

type FieldError struct {
	Field   string
	Message string
}

type ValidationError struct {
	Fields []FieldError
}

func NewFieldError(field, message string) *ValidationError {
	return &ValidationError{Fields: []FieldError{{Field: field, Message: message}}}
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, field := range e.Fields {
		parts = append(parts, field.Field+" "+field.Message)
	}
	return "validation failed: " + strings.Join(parts, ", ")
}
