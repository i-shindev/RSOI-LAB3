package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RentalInProgress = "IN_PROGRESS"
	RentalFinished   = "FINISHED"
	RentalCanceled   = "CANCELED"
)

type Car struct {
	UID                uuid.UUID
	Brand              string
	Model              string
	RegistrationNumber string
	Power              *int
	Type               *string
	Price              int
	Available          bool
}

type CarsPage struct {
	Page          int
	PageSize      int
	TotalElements int
	Items         []Car
}

type ListCarsQuery struct {
	Page    *int
	Size    *int
	ShowAll bool
}

type Payment struct {
	UID    uuid.UUID
	Status string
	Price  int
}

type Rental struct {
	UID        uuid.UUID
	Username   string
	CarUID     uuid.UUID
	PaymentUID uuid.UUID
	DateFrom   time.Time
	DateTo     time.Time
	Status     string
}

type NewRental struct {
	CarUID     uuid.UUID
	PaymentUID uuid.UUID
	DateFrom   time.Time
	DateTo     time.Time
}

type RentalDetails struct {
	Rental  Rental
	Car     Car
	Payment Payment
}

type Booking struct {
	Rental  Rental
	Payment Payment
}

type BookingInput struct {
	CarUID   *uuid.UUID
	DateFrom *time.Time
	DateTo   *time.Time
}

type BookingRequest struct {
	Username string
	CarUID   uuid.UUID
	DateFrom time.Time
	DateTo   time.Time
}

func (r BookingRequest) Days() int {
	return int(r.DateTo.Sub(r.DateFrom) / (24 * time.Hour))
}
