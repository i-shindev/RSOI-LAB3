package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"gateway/src/domain"
	"gateway/src/service"
)

const username = "Test Max"

type fixture struct {
	cars     *mockCars
	rentals  *mockRentals
	payments *mockPayments
	gateway  *service.Gateway
	car      domain.Car
}

func newFixture() *fixture {
	car := domain.Car{
		UID:                uuid.MustParse("109b42f3-198d-4c89-9276-a7520a7120ab"),
		Brand:              "Mercedes Benz",
		Model:              "GLA 250",
		RegistrationNumber: "ЛО777Х799",
		Price:              3500,
		Available:          true,
	}
	f := &fixture{
		cars:     &mockCars{cars: map[uuid.UUID]domain.Car{car.UID: car}},
		rentals:  &mockRentals{rentals: map[uuid.UUID]domain.Rental{}},
		payments: &mockPayments{payments: map[uuid.UUID]domain.Payment{}},
		car:      car,
	}
	f.gateway = service.NewGateway(f.cars, f.rentals, f.payments)
	return f
}

func date(value string) *time.Time {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		panic(err)
	}
	return &parsed
}

func (f *fixture) book(t *testing.T) domain.Booking {
	t.Helper()
	booking, err := f.gateway.BookCar(context.Background(), username, domain.BookingInput{
		CarUID:   &f.car.UID,
		DateFrom: date("2021-10-08"),
		DateTo:   date("2021-10-11"),
	})
	if err != nil {
		t.Fatalf("booking failed: %v", err)
	}
	return booking
}

func TestListCarsPassesQueryToCarsService(t *testing.T) {
	f := newFixture()
	page, size := 1, 10
	query := domain.ListCarsQuery{Page: &page, Size: &size, ShowAll: true}

	result, err := f.gateway.ListCars(context.Background(), query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.cars.lastQuery != query {
		t.Fatalf("cars service got %+v, want %+v", f.cars.lastQuery, query)
	}
	if len(result.Items) != 1 || result.Items[0] != f.car {
		t.Fatalf("expected the seeded car, got %+v", result.Items)
	}
}

func TestListRentalsCombinesCarAndPayment(t *testing.T) {
	f := newFixture()
	booking := f.book(t)

	rentals, err := f.gateway.ListRentals(context.Background(), username)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rentals) != 1 {
		t.Fatalf("expected one rental, got %+v", rentals)
	}
	got := rentals[0]
	if got.Rental.UID != booking.Rental.UID || got.Car.UID != f.car.UID || got.Payment != booking.Payment {
		t.Fatalf("rental is not combined with its car and payment: %+v", got)
	}
}

func TestGetRentalCombinesCarAndPayment(t *testing.T) {
	f := newFixture()
	booking := f.book(t)

	got, err := f.gateway.GetRental(context.Background(), username, booking.Rental.UID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Rental != booking.Rental || got.Car.RegistrationNumber != "ЛО777Х799" || got.Payment != booking.Payment {
		t.Fatalf("rental is not combined with its car and payment: %+v", got)
	}
}

func TestBookCarReservesCarAndChargesForEveryDay(t *testing.T) {
	f := newFixture()

	booking := f.book(t)

	if booking.Payment.Price != 3*3500 || booking.Payment.Status != "PAID" {
		t.Fatalf("expected PAID payment of 10500, got %+v", booking.Payment)
	}
	if booking.Rental.Status != domain.RentalInProgress || booking.Rental.PaymentUID != booking.Payment.UID ||
		booking.Rental.CarUID != f.car.UID || booking.Rental.Username != username {
		t.Fatalf("unexpected rental: %+v", booking.Rental)
	}
	if f.cars.cars[f.car.UID].Available {
		t.Fatal("car should be reserved")
	}
}

func TestFinishRentalReleasesCarAndKeepsPayment(t *testing.T) {
	f := newFixture()
	booking := f.book(t)

	if err := f.gateway.FinishRental(context.Background(), username, booking.Rental.UID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status := f.rentals.rentals[booking.Rental.UID].Status; status != domain.RentalFinished {
		t.Fatalf("expected FINISHED, got %s", status)
	}
	if !f.cars.cars[f.car.UID].Available {
		t.Fatal("car should be released")
	}
	if status := f.payments.payments[booking.Payment.UID].Status; status != "PAID" {
		t.Fatalf("payment must stay PAID, got %s", status)
	}
}

func TestCancelRentalReleasesCarAndCancelsPayment(t *testing.T) {
	f := newFixture()
	booking := f.book(t)

	if err := f.gateway.CancelRental(context.Background(), username, booking.Rental.UID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status := f.rentals.rentals[booking.Rental.UID].Status; status != domain.RentalCanceled {
		t.Fatalf("expected CANCELED, got %s", status)
	}
	if !f.cars.cars[f.car.UID].Available {
		t.Fatal("car should be released")
	}
	if status := f.payments.payments[booking.Payment.UID].Status; status != "CANCELED" {
		t.Fatalf("expected payment CANCELED, got %s", status)
	}
}
