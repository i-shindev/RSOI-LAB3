package api

import "payment-service/src/domain"

type createPaymentRequest struct {
	Price *int `json:"price"`
}

type paymentResponse struct {
	PaymentUID string `json:"paymentUid"`
	Status     string `json:"status"`
	Price      int    `json:"price"`
}

func toPaymentResponse(payment domain.Payment) paymentResponse {
	return paymentResponse{
		PaymentUID: payment.UID.String(),
		Status:     string(payment.Status),
		Price:      payment.Price,
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

type errorDescription struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type validationErrorResponse struct {
	Message string             `json:"message"`
	Errors  []errorDescription `json:"errors"`
}
