package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"cars-service/src/domain"
)

type CarService struct {
	repository domain.CarRepository
}

func NewCarService(repository domain.CarRepository) *CarService {
	return &CarService{repository: repository}
}

func (s *CarService) List(ctx context.Context, input domain.ListCarsInput) (domain.CarsPage, error) {
	query, err := input.Validate()
	if err != nil {
		return domain.CarsPage{}, err
	}
	cars, total, err := s.repository.List(ctx, query.Offset(), query.Size, query.ShowAll)
	if err != nil {
		return domain.CarsPage{}, err
	}
	return domain.CarsPage{Page: query.Page, TotalElements: total, Items: cars}, nil
}

func (s *CarService) Get(ctx context.Context, uid uuid.UUID) (domain.Car, error) {
	car, found, err := s.repository.FindByUID(ctx, uid)
	if err != nil {
		return domain.Car{}, err
	}
	if !found {
		return domain.Car{}, fmt.Errorf("%w: %s", domain.ErrNotFound, uid)
	}
	return car, nil
}

func (s *CarService) Reserve(ctx context.Context, uid uuid.UUID) error {
	if _, err := s.Get(ctx, uid); err != nil {
		return err
	}
	reserved, err := s.repository.TryReserve(ctx, uid)
	if err != nil {
		return err
	}
	if !reserved {
		return fmt.Errorf("%w: %s", domain.ErrAlreadyReserved, uid)
	}
	return nil
}

func (s *CarService) Release(ctx context.Context, uid uuid.UUID) error {
	car, err := s.Get(ctx, uid)
	if err != nil {
		return err
	}
	if car.Available {
		return nil
	}
	return s.repository.Release(ctx, uid)
}
