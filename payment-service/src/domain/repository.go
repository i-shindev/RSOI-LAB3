package domain

import (
	"context"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment Payment) error
	FindByUID(ctx context.Context, uid uuid.UUID) (Payment, bool, error)
	UpdateStatus(ctx context.Context, uid uuid.UUID, status PaymentStatus) error
}
