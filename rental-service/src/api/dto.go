package api

import (
	"time"

	"rental-service/src/domain"
)

const dateLayout = time.DateOnly

type createRentalRequest struct {
	CarUID     *string `json:"carUid"`
	PaymentUID *string `json:"paymentUid"`
	DateFrom   *string `json:"dateFrom"`
	DateTo     *string `json:"dateTo"`
}

type rentalResponse struct {
	RentalUID  string `json:"rentalUid"`
	Username   string `json:"username"`
	CarUID     string `json:"carUid"`
	PaymentUID string `json:"paymentUid"`
	DateFrom   string `json:"dateFrom"`
	DateTo     string `json:"dateTo"`
	Status     string `json:"status"`
}

func toRentalResponse(rental domain.Rental) rentalResponse {
	return rentalResponse{
		RentalUID:  rental.UID.String(),
		Username:   rental.Username,
		CarUID:     rental.CarUID.String(),
		PaymentUID: rental.PaymentUID.String(),
		DateFrom:   rental.DateFrom.UTC().Format(dateLayout),
		DateTo:     rental.DateTo.UTC().Format(dateLayout),
		Status:     string(rental.Status),
	}
}

func toRentalResponses(rentals []domain.Rental) []rentalResponse {
	responses := make([]rentalResponse, 0, len(rentals))
	for _, rental := range rentals {
		responses = append(responses, toRentalResponse(rental))
	}
	return responses
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
