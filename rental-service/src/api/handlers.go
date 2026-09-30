package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"rental-service/src/domain"
	"rental-service/src/service"
)

const usernameHeader = "X-User-Name"

type Handlers struct {
	rentals *service.RentalService
}

func NewHandlers(rentals *service.RentalService) *Handlers {
	return &Handlers{rentals: rentals}
}

func (h *Handlers) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) listRentals(w http.ResponseWriter, r *http.Request) {
	rentals, err := h.rentals.List(r.Context(), r.Header.Get(usernameHeader))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toRentalResponses(rentals))
}

func (h *Handlers) createRental(w http.ResponseWriter, r *http.Request) {
	var request createRentalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeValidationError(w, "Malformed JSON body", nil)
		return
	}
	input, fields := parseNewRentalInput(request)
	if len(fields) > 0 {
		writeValidationError(w, "Validation failed", fields)
		return
	}
	rental, err := h.rentals.Create(r.Context(), r.Header.Get(usernameHeader), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/rentals/"+rental.UID.String())
	writeJSON(w, http.StatusCreated, toRentalResponse(rental))
}

func (h *Handlers) getRental(w http.ResponseWriter, r *http.Request) {
	uid, ok := rentalUID(w, r)
	if !ok {
		return
	}
	rental, err := h.rentals.Get(r.Context(), r.Header.Get(usernameHeader), uid)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toRentalResponse(rental))
}

func (h *Handlers) finishRental(w http.ResponseWriter, r *http.Request) {
	uid, ok := rentalUID(w, r)
	if !ok {
		return
	}
	if err := h.rentals.Finish(r.Context(), r.Header.Get(usernameHeader), uid); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) cancelRental(w http.ResponseWriter, r *http.Request) {
	uid, ok := rentalUID(w, r)
	if !ok {
		return
	}
	if err := h.rentals.Cancel(r.Context(), r.Header.Get(usernameHeader), uid); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseNewRentalInput(request createRentalRequest) (domain.NewRentalInput, []domain.FieldError) {
	var fields []domain.FieldError
	parseUID := func(name string, raw *string) *uuid.UUID {
		if raw == nil {
			return nil
		}
		value, err := uuid.Parse(*raw)
		if err != nil {
			fields = append(fields, domain.FieldError{Field: name, Message: "must be a valid UUID"})
			return nil
		}
		return &value
	}
	parseDate := func(name string, raw *string) *time.Time {
		if raw == nil {
			return nil
		}
		value, err := time.Parse(dateLayout, *raw)
		if err != nil {
			fields = append(fields, domain.FieldError{Field: name, Message: "must be a date in YYYY-MM-DD format"})
			return nil
		}
		return &value
	}
	input := domain.NewRentalInput{
		CarUID:     parseUID("carUid", request.CarUID),
		PaymentUID: parseUID("paymentUid", request.PaymentUID),
		DateFrom:   parseDate("dateFrom", request.DateFrom),
		DateTo:     parseDate("dateTo", request.DateTo),
	}
	return input, fields
}

func rentalUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	uid, err := uuid.Parse(r.PathValue("rentalUid"))
	if err != nil {
		writeValidationError(w, "Validation failed", []domain.FieldError{
			{Field: "rentalUid", Message: "must be a valid UUID"},
		})
		return uuid.Nil, false
	}
	return uid, true
}
