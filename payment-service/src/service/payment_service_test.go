package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"payment-service/src/domain"
	"payment-service/src/service"
)

func price(value int) *int {
	return &value
}

func TestCreateStoresPaidPayment(t *testing.T) {
	repository := newMockPaymentRepository()
	payments := service.NewPaymentService(repository)

	payment, err := payments.Create(context.Background(), domain.NewPaymentInput{Price: price(10500)})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.UID == uuid.Nil || payment.Status != domain.StatusPaid || payment.Price != 10500 {
		t.Fatalf("unexpected payment: %+v", payment)
	}
	if stored := repository.payments[payment.UID]; stored != payment {
		t.Fatalf("stored %+v, returned %+v", stored, payment)
	}
}

func TestCreateRejectsInvalidPrice(t *testing.T) {
	cases := map[string]struct {
		input   domain.NewPaymentInput
		message string
	}{
		"missing":  {domain.NewPaymentInput{}, "must be present"},
		"negative": {domain.NewPaymentInput{Price: price(-1)}, "must not be negative"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			repository := newMockPaymentRepository()
			payments := service.NewPaymentService(repository)

			_, err := payments.Create(context.Background(), tc.input)

			var validation *domain.ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("expected validation error, got %v", err)
			}
			want := domain.FieldError{Field: "price", Message: tc.message}
			if len(validation.Fields) != 1 || validation.Fields[0] != want {
				t.Fatalf("expected %+v, got %+v", want, validation.Fields)
			}
			if len(repository.payments) != 0 {
				t.Fatalf("nothing should be stored, got %+v", repository.payments)
			}
		})
	}
}

func TestGetUnknownPaymentIsNotFound(t *testing.T) {
	payments := service.NewPaymentService(newMockPaymentRepository())

	_, err := payments.Get(context.Background(), uuid.New())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCancelMarksPaymentCanceledAndIsIdempotent(t *testing.T) {
	paid := domain.Payment{UID: uuid.New(), Status: domain.StatusPaid, Price: 10500}
	repository := newMockPaymentRepository(paid)
	payments := service.NewPaymentService(repository)

	for attempt := 1; attempt <= 2; attempt++ {
		if err := payments.Cancel(context.Background(), paid.UID); err != nil {
			t.Fatalf("attempt %d: unexpected error: %v", attempt, err)
		}
	}

	want := domain.Payment{UID: paid.UID, Status: domain.StatusCanceled, Price: 10500}
	if stored := repository.payments[paid.UID]; stored != want {
		t.Fatalf("expected %+v, got %+v", want, stored)
	}
}

func TestCancelUnknownPaymentIsNotFound(t *testing.T) {
	repository := newMockPaymentRepository()
	payments := service.NewPaymentService(repository)

	err := payments.Cancel(context.Background(), uuid.New())

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if len(repository.payments) != 0 {
		t.Fatalf("nothing should be stored, got %+v", repository.payments)
	}
}
