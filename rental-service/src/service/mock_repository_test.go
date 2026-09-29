package service_test

import (
	"context"

	"github.com/google/uuid"

	"rental-service/src/domain"
)

type mockRentalRepository struct {
	rentals []domain.Rental
}

func newMockRentalRepository(rentals ...domain.Rental) *mockRentalRepository {
	return &mockRentalRepository{rentals: rentals}
}

func (r *mockRentalRepository) Create(_ context.Context, rental domain.Rental) error {
	r.rentals = append(r.rentals, rental)
	return nil
}

func (r *mockRentalRepository) ListByUsername(_ context.Context, username string) ([]domain.Rental, error) {
	var owned []domain.Rental
	for _, rental := range r.rentals {
		if rental.Username == username {
			owned = append(owned, rental)
		}
	}
	return owned, nil
}

func (r *mockRentalRepository) FindByUID(_ context.Context, uid uuid.UUID) (domain.Rental, bool, error) {
	if i := r.index(uid); i >= 0 {
		return r.rentals[i], true, nil
	}
	return domain.Rental{}, false, nil
}

func (r *mockRentalRepository) UpdateStatus(_ context.Context, uid uuid.UUID, from, to domain.RentalStatus) (bool, error) {
	i := r.index(uid)
	if i < 0 || r.rentals[i].Status != from {
		return false, nil
	}
	r.rentals[i].Status = to
	return true, nil
}

func (r *mockRentalRepository) index(uid uuid.UUID) int {
	for i, rental := range r.rentals {
		if rental.UID == uid {
			return i
		}
	}
	return -1
}
