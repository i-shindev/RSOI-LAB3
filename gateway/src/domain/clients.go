package domain

import (
	"context"

	"github.com/google/uuid"
)

type CarsClient interface {
	List(ctx context.Context, query ListCarsQuery) (CarsPage, error)
	Get(ctx context.Context, uid uuid.UUID) (Car, error)
	Reserve(ctx context.Context, uid uuid.UUID) error
	Release(ctx context.Context, uid uuid.UUID) error
}

type RentalClient interface {
	List(ctx context.Context, username string) ([]Rental, error)
	Get(ctx context.Context, username string, uid uuid.UUID) (Rental, error)
	Create(ctx context.Context, username string, rental NewRental) (Rental, error)
	Finish(ctx context.Context, username string, uid uuid.UUID) error
	Cancel(ctx context.Context, username string, uid uuid.UUID) error
}

type PaymentClient interface {
	Create(ctx context.Context, price int) (Payment, error)
	Get(ctx context.Context, uid uuid.UUID) (Payment, error)
	Cancel(ctx context.Context, uid uuid.UUID) error
}
