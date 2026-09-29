package service_test

import (
	"context"

	"github.com/google/uuid"

	"cars-service/src/domain"
)

type mockCarRepository struct {
	cars []domain.Car
}

func newMockCarRepository(cars ...domain.Car) *mockCarRepository {
	return &mockCarRepository{cars: cars}
}

func (r *mockCarRepository) List(_ context.Context, offset, limit int, showAll bool) ([]domain.Car, int, error) {
	var visible []domain.Car
	for _, car := range r.cars {
		if showAll || car.Available {
			visible = append(visible, car)
		}
	}
	start := min(offset, len(visible))
	end := min(offset+limit, len(visible))
	return visible[start:end], len(visible), nil
}

func (r *mockCarRepository) FindByUID(_ context.Context, uid uuid.UUID) (domain.Car, bool, error) {
	if i := r.index(uid); i >= 0 {
		return r.cars[i], true, nil
	}
	return domain.Car{}, false, nil
}

func (r *mockCarRepository) TryReserve(_ context.Context, uid uuid.UUID) (bool, error) {
	i := r.index(uid)
	if i < 0 || !r.cars[i].Available {
		return false, nil
	}
	r.cars[i].Available = false
	return true, nil
}

func (r *mockCarRepository) Release(_ context.Context, uid uuid.UUID) error {
	if i := r.index(uid); i >= 0 {
		r.cars[i].Available = true
	}
	return nil
}

func (r *mockCarRepository) index(uid uuid.UUID) int {
	for i, car := range r.cars {
		if car.UID == uid {
			return i
		}
	}
	return -1
}
