package service_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"cars-service/src/domain"
	"cars-service/src/service"
)

func car(brand string, available bool) domain.Car {
	return domain.Car{UID: uuid.New(), Brand: brand, Model: "Model", RegistrationNumber: "A000AA00", Price: 1000, Available: available}
}

func intPtr(value int) *int {
	return &value
}

func brands(cars []domain.Car) []string {
	names := make([]string, 0, len(cars))
	for _, car := range cars {
		names = append(names, car.Brand)
	}
	return names
}

func TestListHidesReservedCarsUnlessShowAll(t *testing.T) {
	cars := service.NewCarService(newMockCarRepository(car("A", true), car("B", false), car("C", true)))

	available, err := cars.List(context.Background(), domain.ListCarsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	all, err := cars.List(context.Background(), domain.ListCarsInput{ShowAll: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := brands(available.Items); !slices.Equal(got, []string{"A", "C"}) || available.TotalElements != 2 {
		t.Fatalf("available: got %v, total %d", got, available.TotalElements)
	}
	if got := brands(all.Items); !slices.Equal(got, []string{"A", "B", "C"}) || all.TotalElements != 3 {
		t.Fatalf("showAll: got %v, total %d", got, all.TotalElements)
	}
}

func TestListReturnsRequestedPageStartingFromOne(t *testing.T) {
	cars := service.NewCarService(newMockCarRepository(car("A", true), car("B", true), car("C", true)))

	page, err := cars.List(context.Background(), domain.ListCarsInput{Page: intPtr(2), Size: intPtr(2)})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := brands(page.Items); page.Page != 2 || page.TotalElements != 3 || !slices.Equal(got, []string{"C"}) {
		t.Fatalf("got page %d, total %d, items %v", page.Page, page.TotalElements, got)
	}
}

func TestListRejectsInvalidPagination(t *testing.T) {
	cars := service.NewCarService(newMockCarRepository())

	_, err := cars.List(context.Background(), domain.ListCarsInput{Page: intPtr(0), Size: intPtr(101)})

	var validation *domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	var fields []string
	for _, field := range validation.Fields {
		fields = append(fields, field.Field)
	}
	if !slices.Equal(fields, []string{"page", "size"}) {
		t.Fatalf("expected errors for page and size, got %+v", validation.Fields)
	}
}

func TestReserveTakesCarOnlyOnce(t *testing.T) {
	free := car("A", true)
	repository := newMockCarRepository(free)
	cars := service.NewCarService(repository)

	if err := cars.Reserve(context.Background(), free.UID); err != nil {
		t.Fatalf("first reservation: unexpected error: %v", err)
	}
	err := cars.Reserve(context.Background(), free.UID)

	if !errors.Is(err, domain.ErrAlreadyReserved) {
		t.Fatalf("second reservation: expected ErrAlreadyReserved, got %v", err)
	}
	if repository.cars[0].Available {
		t.Fatal("car should stay reserved")
	}
}

func TestReserveUnknownCarIsNotFound(t *testing.T) {
	cars := service.NewCarService(newMockCarRepository(car("A", true)))

	err := cars.Reserve(context.Background(), uuid.New())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestReleaseFreesCarAndIsIdempotent(t *testing.T) {
	reserved := car("A", false)
	repository := newMockCarRepository(reserved)
	cars := service.NewCarService(repository)

	for attempt := 1; attempt <= 2; attempt++ {
		if err := cars.Release(context.Background(), reserved.UID); err != nil {
			t.Fatalf("attempt %d: unexpected error: %v", attempt, err)
		}
	}

	if !repository.cars[0].Available {
		t.Fatal("car should be available")
	}
}
