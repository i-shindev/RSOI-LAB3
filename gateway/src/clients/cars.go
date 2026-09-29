package clients

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	"gateway/src/domain"
)

type CarsClient struct {
	client
	circuit_breaker CircuitBreakerStrategy
}

func NewCarsClient(baseURL string, timeout time.Duration) *CarsClient {
	return &CarsClient{client: newClient("Cars Service", baseURL, timeout), circuit_breaker: DefaultCircuitBreaker()}
}

type carDTO struct {
	CarUID             uuid.UUID `json:"carUid"`
	Brand              string    `json:"brand"`
	Model              string    `json:"model"`
	RegistrationNumber string    `json:"registrationNumber"`
	Power              *int      `json:"power"`
	Type               *string   `json:"type"`
	Price              int       `json:"price"`
	Available          bool      `json:"available"`
}

func (d carDTO) toDomain() domain.Car {
	return domain.Car{
		UID:                d.CarUID,
		Brand:              d.Brand,
		Model:              d.Model,
		RegistrationNumber: d.RegistrationNumber,
		Power:              d.Power,
		Type:               d.Type,
		Price:              d.Price,
		Available:          d.Available,
	}
}

type carsPageDTO struct {
	Page          int      `json:"page"`
	PageSize      int      `json:"pageSize"`
	TotalElements int      `json:"totalElements"`
	Items         []carDTO `json:"items"`
}

func (c *CarsClient) List(ctx context.Context, query domain.ListCarsQuery) (domain.CarsPage, error) {
	if IsOpen(c.circuit_breaker) {
		return domain.CarsPage{}, nil
	}
	params := url.Values{"showAll": {strconv.FormatBool(query.ShowAll)}}
	if query.Page != nil {
		params.Set("page", strconv.Itoa(*query.Page))
	}
	if query.Size != nil {
		params.Set("size", strconv.Itoa(*query.Size))
	}
	var page carsPageDTO
	if err := WrapCB(c.circuit_breaker, c.do(ctx, request{method: http.MethodGet, path: "/api/v1/cars", query: params}, &page)); err != nil {
		return domain.CarsPage{}, err
	}
	items := make([]domain.Car, 0, len(page.Items))
	for _, car := range page.Items {
		items = append(items, car.toDomain())
	}
	return domain.CarsPage{
		Page:          page.Page,
		PageSize:      page.PageSize,
		TotalElements: page.TotalElements,
		Items:         items,
	}, nil
}

func (c *CarsClient) Get(ctx context.Context, uid uuid.UUID) (domain.Car, error) {
	var car carDTO
	if IsOpen(c.circuit_breaker) {
		car.CarUID = uid
		return car.toDomain(), nil
	}
	if err := WrapCB(c.circuit_breaker, c.do(ctx, request{method: http.MethodGet, path: "/api/v1/cars/" + uid.String()}, &car)); err != nil {
		return domain.Car{}, err
	}
	return car.toDomain(), nil
}

func (c *CarsClient) Reserve(ctx context.Context, uid uuid.UUID) error {
	return WrapCB(c.circuit_breaker, c.do(ctx, request{method: http.MethodPost, path: "/api/v1/cars/" + uid.String() + "/reservation"}, nil))
}

func (c *CarsClient) Release(ctx context.Context, uid uuid.UUID) error {
	return WrapCB(c.circuit_breaker, c.do(ctx, request{method: http.MethodDelete, path: "/api/v1/cars/" + uid.String() + "/reservation"}, nil))
}
