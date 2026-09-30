package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"gateway/src/domain"
)

type Gateway struct {
	cars     domain.CarsClient
	rentals  domain.RentalClient
	payments domain.PaymentClient
	rq       RetryScheduler
}

type RetryScheduler interface {
	ScheduleRetry(name string, work func(ctx context.Context) error)
}

func NewGateway(cars domain.CarsClient, rentals domain.RentalClient, payments domain.PaymentClient, retry_queue RetryScheduler) *Gateway {
	return &Gateway{cars: cars, rentals: rentals, payments: payments, rq: retry_queue}
}

func (g *Gateway) ListCars(ctx context.Context, query domain.ListCarsQuery) (domain.CarsPage, error) {
	return g.cars.List(ctx, query)
}

func (g *Gateway) ListRentals(ctx context.Context, username string) ([]domain.RentalDetails, error) {
	if err := domain.ValidateUsername(username); err != nil {
		return nil, err
	}
	rentals, err := g.rentals.List(ctx, username)
	if err != nil {
		return nil, err
	}
	cars := map[uuid.UUID]domain.Car{}
	details := make([]domain.RentalDetails, 0, len(rentals))
	for _, rental := range rentals {
		car, cached := cars[rental.CarUID]
		if !cached {
			car, err = g.cars.Get(ctx, rental.CarUID)
			if errors.Is(err, domain.ErrServiceUnavailable) {
				err = nil
				car = domain.Car{UID: rental.CarUID}
			}
			if err != nil {
				return nil, err
			}
			cars[rental.CarUID] = car
		}
		payment, err := g.payments.Get(ctx, rental.PaymentUID)
		if errors.Is(err, domain.ErrServiceUnavailable) {
			err = nil
			payment = domain.Payment{UID: rental.PaymentUID}
		}
		if err != nil {
			return nil, err
		}
		details = append(details, domain.RentalDetails{Rental: rental, Car: car, Payment: payment})
	}
	return details, nil
}

func (g *Gateway) GetRental(ctx context.Context, username string, uid uuid.UUID) (domain.RentalDetails, error) {
	if err := domain.ValidateUsername(username); err != nil {
		return domain.RentalDetails{}, err
	}
	rental, err := g.rentals.Get(ctx, username, uid)
	if err != nil {
		return domain.RentalDetails{}, err
	}
	car, err := g.cars.Get(ctx, rental.CarUID)
	if errors.Is(err, domain.ErrServiceUnavailable) {
		err = nil
		car = domain.Car{UID: rental.CarUID}
	}
	if err != nil {
		return domain.RentalDetails{}, err
	}
	payment, err := g.payments.Get(ctx, rental.PaymentUID)
	if errors.Is(err, domain.ErrServiceUnavailable) {
		err = nil
		payment = domain.Payment{UID: rental.PaymentUID}
	}
	if err != nil {
		return domain.RentalDetails{}, err
	}
	return domain.RentalDetails{Rental: rental, Car: car, Payment: payment}, nil
}

func (g *Gateway) BookCar(ctx context.Context, username string, input domain.BookingInput) (domain.Booking, error) {
	request, err := input.Validate(username)
	if err != nil {
		return domain.Booking{}, err
	}

	car, err := g.cars.Get(ctx, request.CarUID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Booking{}, domain.NewFieldError("carUid", "car not found")
	}
	if err != nil {
		return domain.Booking{}, err
	}
	err = g.cars.Reserve(ctx, car.UID)
	if errors.Is(err, domain.ErrConflict) {
		return domain.Booking{}, domain.NewFieldError("carUid", "car is not available for booking")
	}
	if err != nil {
		return domain.Booking{}, err
	}

	payment, err := g.payments.Create(ctx, request.Days()*car.Price)
	if err != nil {
		g.runOrRetry(ctx, "release car", func(ctx context.Context) error { return g.cars.Release(ctx, car.UID) })
		return domain.Booking{}, err
	}

	rental, err := g.rentals.Create(ctx, username, domain.NewRental{
		CarUID:     car.UID,
		PaymentUID: payment.UID,
		DateFrom:   request.DateFrom,
		DateTo:     request.DateTo,
	})
	if err != nil {
		g.runOrRetry(ctx, "cancel payment", func(ctx context.Context) error { return g.payments.Cancel(ctx, payment.UID) })
		g.runOrRetry(ctx, "release car", func(ctx context.Context) error { return g.cars.Release(ctx, car.UID) })
		return domain.Booking{}, err
	}
	return domain.Booking{Rental: rental, Payment: payment}, nil
}

func (g *Gateway) FinishRental(ctx context.Context, username string, uid uuid.UUID) error {
	rental, err := g.openRental(ctx, username, uid, domain.RentalFinished)
	if err != nil || rental == nil {
		return err
	}
	if err := g.rentals.Finish(ctx, username, uid); err != nil {
		return err
	}
	g.runOrRetry(ctx, "release car", func(ctx context.Context) error { return g.cars.Release(ctx, rental.CarUID) })
	return nil
}

func (g *Gateway) CancelRental(ctx context.Context, username string, uid uuid.UUID) error {
	rental, err := g.openRental(ctx, username, uid, domain.RentalCanceled)
	if err != nil || rental == nil {
		return err
	}
	if err := g.rentals.Cancel(ctx, username, uid); err != nil {
		return err
	}
	g.runOrRetry(ctx, "release car", func(ctx context.Context) error { return g.cars.Release(ctx, rental.CarUID) })
	g.runOrRetry(ctx, "cancel payment", func(ctx context.Context) error { return g.payments.Cancel(ctx, rental.PaymentUID) })
	return nil
}

func (g *Gateway) openRental(ctx context.Context, username string, uid uuid.UUID, target string) (*domain.Rental, error) {
	if err := domain.ValidateUsername(username); err != nil {
		return nil, err
	}
	rental, err := g.rentals.Get(ctx, username, uid)
	if err != nil {
		return nil, err
	}
	switch rental.Status {
	case target:
		return nil, nil
	case domain.RentalInProgress:
		return &rental, nil
	default:
		return nil, domain.NewError(domain.ErrConflict, "rental %s is already %s", uid, rental.Status)
	}
}
func (g *Gateway) runOrRetry(ctx context.Context, action string, step func(context.Context) error) {
	var job func(context.Context) error
	job = func(ctx context.Context) error {
		err := step(ctx)
		if errors.Is(err, domain.ErrServiceUnavailable) {
			g.rq.ScheduleRetry(action, job)
		}
		return err
	}
	if err := job(context.WithoutCancel(ctx)); err != nil && !errors.Is(err, domain.ErrServiceUnavailable) {
		slog.Error("step failed", "action", action, "error", err)
	}
}
