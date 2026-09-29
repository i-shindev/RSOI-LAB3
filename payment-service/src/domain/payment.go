package domain

import "github.com/google/uuid"

type PaymentStatus string

const (
	StatusPaid     PaymentStatus = "PAID"
	StatusCanceled PaymentStatus = "CANCELED"
)

type Payment struct {
	UID    uuid.UUID
	Status PaymentStatus
	Price  int
}

type NewPaymentInput struct {
	Price *int
}

type NewPayment struct {
	Price int
}
