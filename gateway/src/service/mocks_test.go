package service_test

import (
	"context"

	"github.com/google/uuid"

	"gateway/src/domain"
)

var errUnavailable = domain.NewError(domain.ErrServiceUnavailable, "service is unavailable")

type mockRetryQueue struct {
	jobs []func(context.Context) error
}

func (m *mockRetryQueue) ScheduleRetry(_ string, work func(context.Context) error) {
	m.jobs = append(m.jobs, work)
}

func (m *mockRetryQueue) runAll() {
	jobs := m.jobs
	m.jobs = nil
	for _, job := range jobs {
		_ = job(context.Background())
	}
}

type mockCars struct {
	cars        map[uuid.UUID]domain.Car
	lastQuery   domain.ListCarsQuery
	unavailable bool
}

func (m *mockCars) List(_ context.Context, query domain.ListCarsQuery) (domain.CarsPage, error) {
	if m.unavailable {
		return domain.CarsPage{}, errUnavailable
	}
	m.lastQuery = query
	var items []domain.Car
	for _, car := range m.cars {
		if query.ShowAll || car.Available {
			items = append(items, car)
		}
	}
	return domain.CarsPage{Page: 1, PageSize: len(items), TotalElements: len(items), Items: items}, nil
}

func (m *mockCars) Get(_ context.Context, uid uuid.UUID) (domain.Car, error) {
	if m.unavailable {
		return domain.Car{}, errUnavailable
	}
	car, found := m.cars[uid]
	if !found {
		return domain.Car{}, domain.NewError(domain.ErrNotFound, "car not found")
	}
	return car, nil
}

func (m *mockCars) Reserve(_ context.Context, uid uuid.UUID) error {
	car := m.cars[uid]
	if !car.Available {
		return domain.NewError(domain.ErrConflict, "car is already reserved")
	}
	car.Available = false
	m.cars[uid] = car
	return nil
}

func (m *mockCars) Release(_ context.Context, uid uuid.UUID) error {
	if m.unavailable {
		return errUnavailable
	}
	car := m.cars[uid]
	car.Available = true
	m.cars[uid] = car
	return nil
}

type mockRentals struct {
	rentals     map[uuid.UUID]domain.Rental
	unavailable bool
}

func (m *mockRentals) List(_ context.Context, username string) ([]domain.Rental, error) {
	if m.unavailable {
		return nil, errUnavailable
	}
	var owned []domain.Rental
	for _, rental := range m.rentals {
		if rental.Username == username {
			owned = append(owned, rental)
		}
	}
	return owned, nil
}

func (m *mockRentals) Get(_ context.Context, username string, uid uuid.UUID) (domain.Rental, error) {
	if m.unavailable {
		return domain.Rental{}, errUnavailable
	}
	rental, found := m.rentals[uid]
	if !found || rental.Username != username {
		return domain.Rental{}, domain.NewError(domain.ErrNotFound, "rental not found")
	}
	return rental, nil
}

func (m *mockRentals) Create(_ context.Context, username string, rental domain.NewRental) (domain.Rental, error) {
	created := domain.Rental{
		UID:        uuid.New(),
		Username:   username,
		CarUID:     rental.CarUID,
		PaymentUID: rental.PaymentUID,
		DateFrom:   rental.DateFrom,
		DateTo:     rental.DateTo,
		Status:     domain.RentalInProgress,
	}
	m.rentals[created.UID] = created
	return created, nil
}

func (m *mockRentals) Finish(_ context.Context, _ string, uid uuid.UUID) error {
	return m.setStatus(uid, domain.RentalFinished)
}

func (m *mockRentals) Cancel(_ context.Context, _ string, uid uuid.UUID) error {
	return m.setStatus(uid, domain.RentalCanceled)
}

func (m *mockRentals) setStatus(uid uuid.UUID, status string) error {
	rental := m.rentals[uid]
	rental.Status = status
	m.rentals[uid] = rental
	return nil
}

type mockPayments struct {
	payments    map[uuid.UUID]domain.Payment
	unavailable bool
}

func (m *mockPayments) Create(_ context.Context, price int) (domain.Payment, error) {
	if m.unavailable {
		return domain.Payment{}, errUnavailable
	}
	payment := domain.Payment{UID: uuid.New(), Status: "PAID", Price: price}
	m.payments[payment.UID] = payment
	return payment, nil
}

func (m *mockPayments) Get(_ context.Context, uid uuid.UUID) (domain.Payment, error) {
	if m.unavailable {
		return domain.Payment{}, errUnavailable
	}
	payment, found := m.payments[uid]
	if !found {
		return domain.Payment{}, domain.NewError(domain.ErrNotFound, "payment not found")
	}
	return payment, nil
}

func (m *mockPayments) Cancel(_ context.Context, uid uuid.UUID) error {
	if m.unavailable {
		return errUnavailable
	}
	payment := m.payments[uid]
	payment.Status = "CANCELED"
	m.payments[uid] = payment
	return nil
}
