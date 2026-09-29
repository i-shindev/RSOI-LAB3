package clients

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"gateway/src/domain"
)

type RentalClient struct {
	client
	circuit_breaker CircuitBreakerStrategy
}

func NewRentalClient(baseURL string, timeout time.Duration) *RentalClient {
	return &RentalClient{client: newClient("Rental Service", baseURL, timeout), circuit_breaker: DefaultCircuitBreaker()}
}

type rentalDTO struct {
	RentalUID  uuid.UUID `json:"rentalUid"`
	Username   string    `json:"username"`
	CarUID     uuid.UUID `json:"carUid"`
	PaymentUID uuid.UUID `json:"paymentUid"`
	DateFrom   string    `json:"dateFrom"`
	DateTo     string    `json:"dateTo"`
	Status     string    `json:"status"`
}

func (d rentalDTO) toDomain() (domain.Rental, error) {
	dateFrom, err := time.Parse(time.DateOnly, d.DateFrom)
	if err != nil {
		return domain.Rental{}, fmt.Errorf("rental %s: dateFrom: %w", d.RentalUID, err)
	}
	dateTo, err := time.Parse(time.DateOnly, d.DateTo)
	if err != nil {
		return domain.Rental{}, fmt.Errorf("rental %s: dateTo: %w", d.RentalUID, err)
	}
	return domain.Rental{
		UID:        d.RentalUID,
		Username:   d.Username,
		CarUID:     d.CarUID,
		PaymentUID: d.PaymentUID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Status:     d.Status,
	}, nil
}

type createRentalDTO struct {
	CarUID     uuid.UUID `json:"carUid"`
	PaymentUID uuid.UUID `json:"paymentUid"`
	DateFrom   string    `json:"dateFrom"`
	DateTo     string    `json:"dateTo"`
}

func (c *RentalClient) List(ctx context.Context, username string) ([]domain.Rental, error) {
	var dtos []rentalDTO
	if IsOpen(c.circuit_breaker) {
		return make([]domain.Rental, 0), nil
	}
	if err := WrapCB(c.circuit_breaker, c.do(ctx, request{method: http.MethodGet, path: "/api/v1/rentals", username: username}, &dtos)); err != nil {
		return nil, err
	}
	rentals := make([]domain.Rental, 0, len(dtos))
	for _, dto := range dtos {
		rental, err := c.convert(dto)
		if err != nil {
			return nil, err
		}
		rentals = append(rentals, rental)
	}
	return rentals, nil
}

func (c *RentalClient) Get(ctx context.Context, username string, uid uuid.UUID) (domain.Rental, error) {
	var dto rentalDTO
	if IsOpen(c.circuit_breaker) {
		var res domain.Rental
		res.Username = username
		res.UID = uid
		return res, nil
	}
	req := request{method: http.MethodGet, path: "/api/v1/rentals/" + uid.String(), username: username}
	if err := WrapCB(c.circuit_breaker, c.do(ctx, req, &dto)); err != nil {
		return domain.Rental{}, err
	}
	return c.convert(dto)
}

func (c *RentalClient) Create(ctx context.Context, username string, rental domain.NewRental) (domain.Rental, error) {
	var dto rentalDTO
	req := request{
		method:   http.MethodPost,
		path:     "/api/v1/rentals",
		username: username,
		body: createRentalDTO{
			CarUID:     rental.CarUID,
			PaymentUID: rental.PaymentUID,
			DateFrom:   rental.DateFrom.Format(time.DateOnly),
			DateTo:     rental.DateTo.Format(time.DateOnly),
		},
	}
	if err := WrapCB(c.circuit_breaker, c.do(ctx, req, &dto)); err != nil {
		return domain.Rental{}, err
	}
	return c.convert(dto)
}

func (c *RentalClient) Finish(ctx context.Context, username string, uid uuid.UUID) error {
	req := request{method: http.MethodPost, path: "/api/v1/rentals/" + uid.String() + "/finish", username: username}
	return WrapCB(c.circuit_breaker, c.do(ctx, req, nil))
}

func (c *RentalClient) Cancel(ctx context.Context, username string, uid uuid.UUID) error {
	req := request{method: http.MethodDelete, path: "/api/v1/rentals/" + uid.String(), username: username}
	return WrapCB(c.circuit_breaker, c.do(ctx, req, nil))
}

func (c *RentalClient) convert(dto rentalDTO) (domain.Rental, error) {
	rental, err := dto.toDomain()
	if err != nil {
		return domain.Rental{}, c.unavailable(err)
	}
	return rental, nil
}
