package domain

import "strings"

const UsernameField = "X-User-Name"

func ValidateUsername(username string) error {
	if strings.TrimSpace(username) == "" {
		return NewFieldError(UsernameField, "must be present")
	}
	return nil
}

func (in BookingInput) Validate(username string) (BookingRequest, error) {
	var fields []FieldError
	missing := func(name string) {
		fields = append(fields, FieldError{Field: name, Message: "must be present"})
	}
	if strings.TrimSpace(username) == "" {
		missing(UsernameField)
	}
	if in.CarUID == nil {
		missing("carUid")
	}
	if in.DateFrom == nil {
		missing("dateFrom")
	}
	if in.DateTo == nil {
		missing("dateTo")
	}
	if in.DateFrom != nil && in.DateTo != nil && !in.DateFrom.Before(*in.DateTo) {
		fields = append(fields, FieldError{Field: "dateTo", Message: "must be after dateFrom"})
	}
	if len(fields) > 0 {
		return BookingRequest{}, &ValidationError{Fields: fields}
	}
	return BookingRequest{
		Username: username,
		CarUID:   *in.CarUID,
		DateFrom: *in.DateFrom,
		DateTo:   *in.DateTo,
	}, nil
}
