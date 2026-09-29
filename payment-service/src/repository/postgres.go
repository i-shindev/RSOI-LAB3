package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"payment-service/src/domain"
)

type PgPaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPgPaymentRepository(pool *pgxpool.Pool) *PgPaymentRepository {
	return &PgPaymentRepository{pool: pool}
}

func (r *PgPaymentRepository) Create(ctx context.Context, payment domain.Payment) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO payment (payment_uid, status, price) VALUES ($1, $2, $3)`,
		payment.UID, string(payment.Status), payment.Price,
	)
	return err
}

func (r *PgPaymentRepository) FindByUID(ctx context.Context, uid uuid.UUID) (domain.Payment, bool, error) {
	var payment domain.Payment
	var status string
	err := r.pool.QueryRow(ctx,
		`SELECT payment_uid, status, price FROM payment WHERE payment_uid = $1`,
		uid,
	).Scan(&payment.UID, &status, &payment.Price)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Payment{}, false, nil
	}
	if err != nil {
		return domain.Payment{}, false, err
	}
	payment.Status = domain.PaymentStatus(status)
	return payment, true, nil
}

func (r *PgPaymentRepository) UpdateStatus(ctx context.Context, uid uuid.UUID, status domain.PaymentStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment SET status = $2 WHERE payment_uid = $1`,
		uid, string(status),
	)
	return err
}
