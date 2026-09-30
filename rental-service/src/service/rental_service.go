package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"rental-service/src/domain"
)

type RentalService struct {
	repository domain.RentalRepository
}

func NewRentalService(repository domain.RentalRepository) *RentalService {
	return &RentalService{repository: repository}
}

func (s *RentalService) List(ctx context.Context, username string) ([]domain.Rental, error) {
	if err := domain.ValidateUsername(username); err != nil {
		return nil, err
	}
	return s.repository.ListByUsername(ctx, username)
}

func (s *RentalService) Get(ctx context.Context, username string, uid uuid.UUID) (domain.Rental, error) {
	if err := domain.ValidateUsername(username); err != nil {
		return domain.Rental{}, err
	}
	rental, found, err := s.repository.FindByUID(ctx, uid)
	if err != nil {
		return domain.Rental{}, err
	}
	if !found || !rental.OwnedBy(username) {
		return domain.Rental{}, fmt.Errorf("%w: %s", domain.ErrNotFound, uid)
	}
	return rental, nil
}

func (s *RentalService) Create(ctx context.Context, username string, input domain.NewRentalInput) (domain.Rental, error) {
	newRental, err := input.Validate(username)
	if err != nil {
		return domain.Rental{}, err
	}
	rental := domain.Rental{
		UID:        uuid.New(),
		Username:   newRental.Username,
		PaymentUID: newRental.PaymentUID,
		CarUID:     newRental.CarUID,
		DateFrom:   newRental.DateFrom,
		DateTo:     newRental.DateTo,
		Status:     domain.StatusInProgress,
	}
	if err := s.repository.Create(ctx, rental); err != nil {
		return domain.Rental{}, err
	}
	return rental, nil
}

func (s *RentalService) Finish(ctx context.Context, username string, uid uuid.UUID) error {
	return s.changeStatus(ctx, username, uid, domain.StatusFinished)
}

func (s *RentalService) Cancel(ctx context.Context, username string, uid uuid.UUID) error {
	return s.changeStatus(ctx, username, uid, domain.StatusCanceled)
}

func (s *RentalService) changeStatus(ctx context.Context, username string, uid uuid.UUID, target domain.RentalStatus) error {
	rental, err := s.Get(ctx, username, uid)
	if err != nil {
		return err
	}
	change, err := rental.Status.CanChangeTo(target)
	if err != nil || !change {
		return err
	}
	changed, err := s.repository.UpdateStatus(ctx, uid, rental.Status, target)
	if err != nil {
		return err
	}
	if !changed {
		return fmt.Errorf("%w: rental %s was changed concurrently", domain.ErrInvalidTransition, uid)
	}
	return nil
}
