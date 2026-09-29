package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RentalStatus string

const (
	StatusInProgress RentalStatus = "IN_PROGRESS"
	StatusFinished   RentalStatus = "FINISHED"
	StatusCanceled   RentalStatus = "CANCELED"
)

func (s RentalStatus) CanChangeTo(target RentalStatus) (bool, error) {
	if s == target {
		return false, nil
	}
	if s != StatusInProgress {
		return false, fmt.Errorf("%w: rental is already %s", ErrInvalidTransition, s)
	}
	return true, nil
}

type Rental struct {
	UID        uuid.UUID
	Username   string
	PaymentUID uuid.UUID
	CarUID     uuid.UUID
	DateFrom   time.Time
	DateTo     time.Time
	Status     RentalStatus
}

func (r Rental) OwnedBy(username string) bool {
	return r.Username == username
}

type NewRentalInput struct {
	CarUID     *uuid.UUID
	PaymentUID *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
}

type NewRental struct {
	Username   string
	CarUID     uuid.UUID
	PaymentUID uuid.UUID
	DateFrom   time.Time
	DateTo     time.Time
}
