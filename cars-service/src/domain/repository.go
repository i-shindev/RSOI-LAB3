package domain

import (
	"context"

	"github.com/google/uuid"
)

type CarRepository interface {
	List(ctx context.Context, offset, limit int, showAll bool) ([]Car, int, error)
	FindByUID(ctx context.Context, uid uuid.UUID) (Car, bool, error)
	TryReserve(ctx context.Context, uid uuid.UUID) (bool, error)
	Release(ctx context.Context, uid uuid.UUID) error
}
