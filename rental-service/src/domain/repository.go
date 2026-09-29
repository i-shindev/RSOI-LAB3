package domain

import (
	"context"

	"github.com/google/uuid"
)

type RentalRepository interface {
	Create(ctx context.Context, rental Rental) error
	ListByUsername(ctx context.Context, username string) ([]Rental, error)
	FindByUID(ctx context.Context, uid uuid.UUID) (Rental, bool, error)
	UpdateStatus(ctx context.Context, uid uuid.UUID, from, to RentalStatus) (bool, error)
}
