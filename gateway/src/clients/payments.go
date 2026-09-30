package clients

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"gateway/src/domain"
)

type PaymentClient struct {
	client
	circuit_breaker CircuitBreakerStrategy
}

func NewPaymentClient(baseURL string, timeout time.Duration) *PaymentClient {
	return &PaymentClient{client: newClient("Payment Service", baseURL, timeout)}
}

type paymentDTO struct {
	PaymentUID uuid.UUID `json:"paymentUid"`
	Status     string    `json:"status"`
	Price      int       `json:"price"`
}

func (d paymentDTO) toDomain() domain.Payment {
	return domain.Payment{UID: d.PaymentUID, Status: d.Status, Price: d.Price}
}

type createPaymentDTO struct {
	Price int `json:"price"`
}

func (c *PaymentClient) Create(ctx context.Context, price int) (domain.Payment, error) {
	var payment paymentDTO
	req := request{method: http.MethodPost, path: "/api/v1/payments", body: createPaymentDTO{Price: price}}
	if err := c.do(ctx, req, &payment); err != nil {
		return domain.Payment{}, err
	}
	return payment.toDomain(), nil
}

func (c *PaymentClient) Get(ctx context.Context, uid uuid.UUID) (domain.Payment, error) {
	var payment paymentDTO
	if err := c.do(ctx, request{method: http.MethodGet, path: "/api/v1/payments/" + uid.String()}, &payment); err != nil {
		return domain.Payment{}, err
	}
	return payment.toDomain(), nil
}

func (c *PaymentClient) Cancel(ctx context.Context, uid uuid.UUID) error {
	return c.do(ctx, request{method: http.MethodDelete, path: "/api/v1/payments/" + uid.String()}, nil)
}
