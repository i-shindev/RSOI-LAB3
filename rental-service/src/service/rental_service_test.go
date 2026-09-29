package service_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"rental-service/src/domain"
	"rental-service/src/service"
)

const owner = "Test Max"

func date(value string) *time.Time {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		panic(err)
	}
	return &parsed
}

func uidPtr(value uuid.UUID) *uuid.UUID {
	return &value
}

func rental(username string, status domain.RentalStatus) domain.Rental {
	return domain.Rental{
		UID:        uuid.New(),
		Username:   username,
		PaymentUID: uuid.New(),
		CarUID:     uuid.New(),
		DateFrom:   *date("2021-10-08"),
		DateTo:     *date("2021-10-11"),
		Status:     status,
	}
}

func TestCreateStoresRentalInProgress(t *testing.T) {
	repository := newMockRentalRepository()
	rentals := service.NewRentalService(repository)
	input := domain.NewRentalInput{
		CarUID:     uidPtr(uuid.New()),
		PaymentUID: uidPtr(uuid.New()),
		DateFrom:   date("2021-10-08"),
		DateTo:     date("2021-10-11"),
	}

	created, err := rentals.Create(context.Background(), owner, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := domain.Rental{
		UID:        created.UID,
		Username:   owner,
		PaymentUID: *input.PaymentUID,
		CarUID:     *input.CarUID,
		DateFrom:   *input.DateFrom,
		DateTo:     *input.DateTo,
		Status:     domain.StatusInProgress,
	}
	if created.UID == uuid.Nil || created != want {
		t.Fatalf("expected %+v, got %+v", want, created)
	}
	if len(repository.rentals) != 1 || repository.rentals[0] != created {
		t.Fatalf("expected the rental to be stored, got %+v", repository.rentals)
	}
}

func TestCreateReportsEveryInvalidField(t *testing.T) {
	repository := newMockRentalRepository()
	rentals := service.NewRentalService(repository)
	input := domain.NewRentalInput{DateFrom: date("2021-10-11"), DateTo: date("2021-10-08")}

	_, err := rentals.Create(context.Background(), "", input)

	var validation *domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	var fields []string
	for _, field := range validation.Fields {
		fields = append(fields, field.Field)
	}
	if want := []string{"X-User-Name", "carUid", "paymentUid", "dateTo"}; !slices.Equal(fields, want) {
		t.Fatalf("expected errors for %v, got %+v", want, validation.Fields)
	}
	if len(repository.rentals) != 0 {
		t.Fatalf("nothing should be stored, got %+v", repository.rentals)
	}
}

func TestListReturnsOnlyOwnRentals(t *testing.T) {
	own := rental(owner, domain.StatusInProgress)
	rentals := service.NewRentalService(newMockRentalRepository(own, rental("Someone Else", domain.StatusInProgress)))

	listed, err := rentals.List(context.Background(), owner)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(listed) != 1 || listed[0] != own {
		t.Fatalf("expected only %+v, got %+v", own, listed)
	}
}

func TestForeignRentalIsNotFound(t *testing.T) {
	foreign := rental("Someone Else", domain.StatusInProgress)
	repository := newMockRentalRepository(foreign)
	rentals := service.NewRentalService(repository)

	_, getErr := rentals.Get(context.Background(), owner, foreign.UID)
	cancelErr := rentals.Cancel(context.Background(), owner, foreign.UID)

	if !errors.Is(getErr, domain.ErrNotFound) || !errors.Is(cancelErr, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got get=%v cancel=%v", getErr, cancelErr)
	}
	if repository.rentals[0].Status != domain.StatusInProgress {
		t.Fatalf("foreign rental must stay untouched, got %s", repository.rentals[0].Status)
	}
}

func TestFinishAndCancelCloseRentalIdempotently(t *testing.T) {
	cases := map[string]struct {
		change func(*service.RentalService, context.Context, string, uuid.UUID) error
		want   domain.RentalStatus
	}{
		"finish": {(*service.RentalService).Finish, domain.StatusFinished},
		"cancel": {(*service.RentalService).Cancel, domain.StatusCanceled},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			active := rental(owner, domain.StatusInProgress)
			repository := newMockRentalRepository(active)
			rentals := service.NewRentalService(repository)

			for attempt := 1; attempt <= 2; attempt++ {
				if err := tc.change(rentals, context.Background(), owner, active.UID); err != nil {
					t.Fatalf("attempt %d: unexpected error: %v", attempt, err)
				}
			}

			if got := repository.rentals[0].Status; got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestClosedRentalCannotSwitchToAnotherFinalStatus(t *testing.T) {
	cases := map[string]struct {
		from   domain.RentalStatus
		change func(*service.RentalService, context.Context, string, uuid.UUID) error
	}{
		"finish canceled": {domain.StatusCanceled, (*service.RentalService).Finish},
		"cancel finished": {domain.StatusFinished, (*service.RentalService).Cancel},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			closed := rental(owner, tc.from)
			repository := newMockRentalRepository(closed)
			rentals := service.NewRentalService(repository)

			err := tc.change(rentals, context.Background(), owner, closed.UID)

			if !errors.Is(err, domain.ErrInvalidTransition) {
				t.Fatalf("expected ErrInvalidTransition, got %v", err)
			}
			if got := repository.rentals[0].Status; got != tc.from {
				t.Fatalf("status must stay %s, got %s", tc.from, got)
			}
		})
	}
}
