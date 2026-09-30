package domain

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	UsernameField     = "X-User-Name"
	MaxUsernameLength = 80
)

func usernameErrors(username string) []FieldError {
	switch {
	case strings.TrimSpace(username) == "":
		return []FieldError{{Field: UsernameField, Message: "must be present"}}
	case utf8.RuneCountInString(username) > MaxUsernameLength:
		return []FieldError{{
			Field:   UsernameField,
			Message: "must be at most " + strconv.Itoa(MaxUsernameLength) + " characters",
		}}
	}
	return nil
}

func ValidateUsername(username string) error {
	if fields := usernameErrors(username); len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}

func (in NewRentalInput) Validate(username string) (NewRental, error) {
	fields := usernameErrors(username)
	missing := func(name string) {
		fields = append(fields, FieldError{Field: name, Message: "must be present"})
	}
	if in.CarUID == nil {
		missing("carUid")
	}
	if in.PaymentUID == nil {
		missing("paymentUid")
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
		return NewRental{}, &ValidationError{Fields: fields}
	}
	return NewRental{
		Username:   username,
		CarUID:     *in.CarUID,
		PaymentUID: *in.PaymentUID,
		DateFrom:   *in.DateFrom,
		DateTo:     *in.DateTo,
	}, nil
}
