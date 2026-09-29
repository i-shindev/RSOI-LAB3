package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"payment-service/src/domain"
)

type PaymentService struct {
	repository domain.PaymentRepository
}

func NewPaymentService(repository domain.PaymentRepository) *PaymentService {
	return &PaymentService{repository: repository}
}

func (s *PaymentService) Create(ctx context.Context, input domain.NewPaymentInput) (domain.Payment, error) {
	newPayment, err := input.Validate()
	if err != nil {
		return domain.Payment{}, err
	}
	payment := domain.Payment{
		UID:    uuid.New(),
		Status: domain.StatusPaid,
		Price:  newPayment.Price,
	}
	if err := s.repository.Create(ctx, payment); err != nil {
		return domain.Payment{}, err
	}
	return payment, nil
}

func (s *PaymentService) Get(ctx context.Context, uid uuid.UUID) (domain.Payment, error) {
	payment, found, err := s.repository.FindByUID(ctx, uid)
	if err != nil {
		return domain.Payment{}, err
	}
	if !found {
		return domain.Payment{}, fmt.Errorf("%w: %s", domain.ErrNotFound, uid)
	}
	return payment, nil
}

func (s *PaymentService) Cancel(ctx context.Context, uid uuid.UUID) error {
	payment, err := s.Get(ctx, uid)
	if err != nil {
		return err
	}
	if payment.Status == domain.StatusCanceled {
		return nil
	}
	return s.repository.UpdateStatus(ctx, uid, domain.StatusCanceled)
}
