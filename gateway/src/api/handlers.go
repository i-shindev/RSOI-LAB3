package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	"gateway/src/domain"
	"gateway/src/service"
)

const usernameHeader = "X-User-Name"

type Handlers struct {
	gateway *service.Gateway
}

func NewHandlers(gateway *service.Gateway) *Handlers {
	return &Handlers{gateway: gateway}
}

func (h *Handlers) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) listCars(w http.ResponseWriter, r *http.Request) {
	query, fields := parseListCarsQuery(r.URL.Query())
	if len(fields) > 0 {
		writeValidationError(w, "Validation failed", fields)
		return
	}
	page, err := h.gateway.ListCars(r.Context(), query)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPaginationResponse(page))
}

func (h *Handlers) listRentals(w http.ResponseWriter, r *http.Request) {
	rentals, err := h.gateway.ListRentals(r.Context(), r.Header.Get(usernameHeader))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toRentalResponses(rentals))
}

func (h *Handlers) getRental(w http.ResponseWriter, r *http.Request) {
	uid, ok := rentalUID(w, r)
	if !ok {
		return
	}
	rental, err := h.gateway.GetRental(r.Context(), r.Header.Get(usernameHeader), uid)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toRentalResponse(rental))
}

func (h *Handlers) bookCar(w http.ResponseWriter, r *http.Request) {
	var request createRentalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeValidationError(w, "Malformed JSON body", nil)
		return
	}
	input, fields := parseBookingInput(request)
	if len(fields) > 0 {
		writeValidationError(w, "Validation failed", fields)
		return
	}
	booking, err := h.gateway.BookCar(r.Context(), r.Header.Get(usernameHeader), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toCreateRentalResponse(booking))
}

func (h *Handlers) finishRental(w http.ResponseWriter, r *http.Request) {
	uid, ok := rentalUID(w, r)
	if !ok {
		return
	}
	if err := h.gateway.FinishRental(r.Context(), r.Header.Get(usernameHeader), uid); err != nil {
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
	if err := h.gateway.CancelRental(r.Context(), r.Header.Get(usernameHeader), uid); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseListCarsQuery(query url.Values) (domain.ListCarsQuery, []domain.FieldError) {
	var result domain.ListCarsQuery
	var fields []domain.FieldError
	parseInt := func(name string) *int {
		raw := query.Get(name)
		if raw == "" {
			return nil
		}
		value, err := strconv.Atoi(raw)
		if err != nil {
			fields = append(fields, domain.FieldError{Field: name, Message: "must be an integer"})
			return nil
		}
		return &value
	}
	result.Page = parseInt("page")
	result.Size = parseInt("size")
	if raw := query.Get("showAll"); raw != "" {
		showAll, err := strconv.ParseBool(raw)
		if err != nil {
			fields = append(fields, domain.FieldError{Field: "showAll", Message: "must be true or false"})
		}
		result.ShowAll = showAll
	}
	return result, fields
}

func parseBookingInput(request createRentalRequest) (domain.BookingInput, []domain.FieldError) {
	var fields []domain.FieldError
	var input domain.BookingInput
	if request.CarUID != nil {
		if uid, err := uuid.Parse(*request.CarUID); err == nil {
			input.CarUID = &uid
		} else {
			fields = append(fields, domain.FieldError{Field: "carUid", Message: "must be a valid UUID"})
		}
	}
	parseDate := func(name string, raw *string) *time.Time {
		if raw == nil {
			return nil
		}
		value, err := time.Parse(time.DateOnly, *raw)
		if err != nil {
			fields = append(fields, domain.FieldError{Field: name, Message: "must be a date in YYYY-MM-DD format"})
			return nil
		}
		return &value
	}
	input.DateFrom = parseDate("dateFrom", request.DateFrom)
	input.DateTo = parseDate("dateTo", request.DateTo)
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
