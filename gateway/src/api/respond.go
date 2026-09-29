package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"gateway/src/domain"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write response", "error", err)
	}
}

func writeValidationError(w http.ResponseWriter, message string, fields []domain.FieldError) {
	errs := make([]errorDescription, 0, len(fields))
	for _, field := range fields {
		errs = append(errs, errorDescription{Field: field.Field, Error: field.Message})
	}
	writeJSON(w, http.StatusBadRequest, validationErrorResponse{Message: message, Errors: errs})
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var validation *domain.ValidationError
	var failure *domain.Error
	message := err.Error()
	if errors.As(err, &failure) {
		message = failure.Message
	}
	switch {
	case errors.As(err, &validation):
		writeValidationError(w, "Validation failed", validation.Fields)
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Message: message})
	case errors.Is(err, domain.ErrConflict):
		writeJSON(w, http.StatusBadRequest, errorResponse{Message: message})
	case errors.Is(err, domain.ErrServiceUnavailable):
		slog.Error("upstream failure", "method", r.Method, "path", r.URL.Path, "error", err)
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Message: message})
	default:
		slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Message: "Internal server error"})
	}
}
