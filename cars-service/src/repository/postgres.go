package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cars-service/src/domain"
)

const carColumns = `car_uid, brand, model, registration_number, power, price, type, availability`

type PgCarRepository struct {
	pool *pgxpool.Pool
}

func NewPgCarRepository(pool *pgxpool.Pool) *PgCarRepository {
	return &PgCarRepository{pool: pool}
}

func (r *PgCarRepository) List(ctx context.Context, offset, limit int, showAll bool) ([]domain.Car, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM cars WHERE $1 OR availability`,
		showAll,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+carColumns+` FROM cars WHERE $1 OR availability ORDER BY id LIMIT $2 OFFSET $3`,
		showAll, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	cars, err := pgx.CollectRows(rows, scanCar)
	if err != nil {
		return nil, 0, err
	}
	return cars, total, nil
}

func (r *PgCarRepository) FindByUID(ctx context.Context, uid uuid.UUID) (domain.Car, bool, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+carColumns+` FROM cars WHERE car_uid = $1`, uid)
	if err != nil {
		return domain.Car{}, false, err
	}
	car, err := pgx.CollectExactlyOneRow(rows, scanCar)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Car{}, false, nil
	}
	if err != nil {
		return domain.Car{}, false, err
	}
	return car, true, nil
}

func (r *PgCarRepository) TryReserve(ctx context.Context, uid uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE cars SET availability = false WHERE car_uid = $1 AND availability`,
		uid,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *PgCarRepository) Release(ctx context.Context, uid uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE cars SET availability = true WHERE car_uid = $1`, uid)
	return err
}

func scanCar(row pgx.CollectableRow) (domain.Car, error) {
	var car domain.Car
	var carType *string
	if err := row.Scan(
		&car.UID,
		&car.Brand,
		&car.Model,
		&car.RegistrationNumber,
		&car.Power,
		&car.Price,
		&carType,
		&car.Available,
	); err != nil {
		return domain.Car{}, err
	}
	if carType != nil {
		t := domain.CarType(*carType)
		car.Type = &t
	}
	return car, nil
}
