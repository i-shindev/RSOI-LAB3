package service_test

import (
	"context"

	"github.com/google/uuid"

	"payment-service/src/domain"
)

type mockPaymentRepository struct {
	payments map[uuid.UUID]domain.Payment
}

func newMockPaymentRepository(payments ...domain.Payment) *mockPaymentRepository {
	repository := &mockPaymentRepository{payments: map[uuid.UUID]domain.Payment{}}
	for _, payment := range payments {
		repository.payments[payment.UID] = payment
	}
	return repository
}

func (r *mockPaymentRepository) Create(_ context.Context, payment domain.Payment) error {
	r.payments[payment.UID] = payment
	return nil
}

func (r *mockPaymentRepository) FindByUID(_ context.Context, uid uuid.UUID) (domain.Payment, bool, error) {
	payment, found := r.payments[uid]
	return payment, found, nil
}

func (r *mockPaymentRepository) UpdateStatus(_ context.Context, uid uuid.UUID, status domain.PaymentStatus) error {
	payment := r.payments[uid]
	payment.Status = status
	r.payments[uid] = payment
	return nil
}
