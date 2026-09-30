package domain

import (
	"errors"
	"strings"
)

var (
	ErrNotFound          = errors.New("rental not found")
	ErrInvalidTransition = errors.New("rental status cannot be changed")
)

type FieldError struct {
	Field   string
	Message string
}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, field := range e.Fields {
		parts = append(parts, field.Field+" "+field.Message)
	}
	return "validation failed: " + strings.Join(parts, ", ")
}
