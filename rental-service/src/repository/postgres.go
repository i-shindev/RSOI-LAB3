package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"rental-service/src/domain"
)

const rentalColumns = `rental_uid, username, payment_uid, car_uid, date_from, date_to, status`

type PgRentalRepository struct {
	pool *pgxpool.Pool
}

func NewPgRentalRepository(pool *pgxpool.Pool) *PgRentalRepository {
	return &PgRentalRepository{pool: pool}
}

func (r *PgRentalRepository) Create(ctx context.Context, rental domain.Rental) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rental (`+rentalColumns+`) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		rental.UID, rental.Username, rental.PaymentUID, rental.CarUID,
		rental.DateFrom, rental.DateTo, string(rental.Status),
	)
	return err
}

func (r *PgRentalRepository) ListByUsername(ctx context.Context, username string) ([]domain.Rental, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+rentalColumns+` FROM rental WHERE username = $1 ORDER BY id`,
		username,
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanRental)
}

func (r *PgRentalRepository) FindByUID(ctx context.Context, uid uuid.UUID) (domain.Rental, bool, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+rentalColumns+` FROM rental WHERE rental_uid = $1`, uid)
	if err != nil {
		return domain.Rental{}, false, err
	}
	rental, err := pgx.CollectExactlyOneRow(rows, scanRental)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Rental{}, false, nil
	}
	if err != nil {
		return domain.Rental{}, false, err
	}
	return rental, true, nil
}

func (r *PgRentalRepository) UpdateStatus(ctx context.Context, uid uuid.UUID, from, to domain.RentalStatus) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE rental SET status = $3 WHERE rental_uid = $1 AND status = $2`,
		uid, string(from), string(to),
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func scanRental(row pgx.CollectableRow) (domain.Rental, error) {
	var rental domain.Rental
	var status string
	if err := row.Scan(
		&rental.UID,
		&rental.Username,
		&rental.PaymentUID,
		&rental.CarUID,
		&rental.DateFrom,
		&rental.DateTo,
		&status,
	); err != nil {
		return domain.Rental{}, err
	}
	rental.Status = domain.RentalStatus(status)
	return rental, nil
}
