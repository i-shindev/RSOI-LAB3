package api

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/uuid"

	"cars-service/src/domain"
	"cars-service/src/service"
)

type Handlers struct {
	cars *service.CarService
}

func NewHandlers(cars *service.CarService) *Handlers {
	return &Handlers{cars: cars}
}

func (h *Handlers) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) listCars(w http.ResponseWriter, r *http.Request) {
	input, fields := parseListCarsInput(r.URL.Query())
	if len(fields) > 0 {
		writeValidationError(w, "Validation failed", fields)
		return
	}
	page, err := h.cars.List(r.Context(), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPaginationResponse(page))
}

func (h *Handlers) getCar(w http.ResponseWriter, r *http.Request) {
	uid, ok := carUID(w, r)
	if !ok {
		return
	}
	car, err := h.cars.Get(r.Context(), uid)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toCarResponse(car))
}

func (h *Handlers) reserveCar(w http.ResponseWriter, r *http.Request) {
	uid, ok := carUID(w, r)
	if !ok {
		return
	}
	if err := h.cars.Reserve(r.Context(), uid); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) releaseCar(w http.ResponseWriter, r *http.Request) {
	uid, ok := carUID(w, r)
	if !ok {
		return
	}
	if err := h.cars.Release(r.Context(), uid); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseListCarsInput(query url.Values) (domain.ListCarsInput, []domain.FieldError) {
	var input domain.ListCarsInput
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
	input.Page = parseInt("page")
	input.Size = parseInt("size")
	if raw := query.Get("showAll"); raw != "" {
		showAll, err := strconv.ParseBool(raw)
		if err != nil {
			fields = append(fields, domain.FieldError{Field: "showAll", Message: "must be true or false"})
		}
		input.ShowAll = showAll
	}
	return input, fields
}

func carUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	uid, err := uuid.Parse(r.PathValue("carUid"))
	if err != nil {
		writeValidationError(w, "Validation failed", []domain.FieldError{
			{Field: "carUid", Message: "must be a valid UUID"},
		})
		return uuid.Nil, false
	}
	return uid, true
}
