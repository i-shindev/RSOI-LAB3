package domain

import (
	"errors"
	"strings"
)

var (
	ErrNotFound        = errors.New("car not found")
	ErrAlreadyReserved = errors.New("car is already reserved")
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
