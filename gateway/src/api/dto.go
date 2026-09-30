package api

import (
	"time"

	"gateway/src/domain"
)

type createRentalRequest struct {
	CarUID   *string `json:"carUid"`
	DateFrom *string `json:"dateFrom"`
	DateTo   *string `json:"dateTo"`
}

type carResponse struct {
	CarUID             string  `json:"carUid"`
	Brand              string  `json:"brand"`
	Model              string  `json:"model"`
	RegistrationNumber string  `json:"registrationNumber"`
	Power              *int    `json:"power,omitempty"`
	Type               *string `json:"type,omitempty"`
	Price              int     `json:"price"`
	Available          bool    `json:"available"`
}

type paginationResponse struct {
	Page          int           `json:"page"`
	PageSize      int           `json:"pageSize"`
	TotalElements int           `json:"totalElements"`
	Items         []carResponse `json:"items"`
}

type carInfo struct {
	CarUID             string `json:"carUid"`
	Brand              string `json:"brand"`
	Model              string `json:"model"`
	RegistrationNumber string `json:"registrationNumber"`
}

type paymentInfo struct {
	PaymentUID string `json:"paymentUid,omitempty"`
	Status     string `json:"status,omitempty"`
	Price      *int   `json:"price,omitempty"`
}

type rentalResponse struct {
	RentalUID string      `json:"rentalUid"`
	Status    string      `json:"status"`
	DateFrom  string      `json:"dateFrom"`
	DateTo    string      `json:"dateTo"`
	Car       carInfo     `json:"car"`
	Payment   paymentInfo `json:"payment"`
}

type createRentalResponse struct {
	RentalUID string      `json:"rentalUid"`
	Status    string      `json:"status"`
	CarUID    string      `json:"carUid"`
	DateFrom  string      `json:"dateFrom"`
	DateTo    string      `json:"dateTo"`
	Payment   paymentInfo `json:"payment"`
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

func formatDate(value time.Time) string {
	return value.UTC().Format(time.DateOnly)
}

func toPaginationResponse(page domain.CarsPage) paginationResponse {
	items := make([]carResponse, 0, len(page.Items))
	for _, car := range page.Items {
		items = append(items, carResponse{
			CarUID:             car.UID.String(),
			Brand:              car.Brand,
			Model:              car.Model,
			RegistrationNumber: car.RegistrationNumber,
			Power:              car.Power,
			Type:               car.Type,
			Price:              car.Price,
			Available:          car.Available,
		})
	}
	return paginationResponse{
		Page:          page.Page,
		PageSize:      page.PageSize,
		TotalElements: page.TotalElements,
		Items:         items,
	}
}

func toPaymentInfo(payment domain.Payment) paymentInfo {
	if payment.Status == "" { // fallback: Payment Service was unavailable
		return paymentInfo{}
	}
	return paymentInfo{PaymentUID: payment.UID.String(), Status: payment.Status, Price: &payment.Price}
}

func toRentalResponse(details domain.RentalDetails) rentalResponse {
	return rentalResponse{
		RentalUID: details.Rental.UID.String(),
		Status:    details.Rental.Status,
		DateFrom:  formatDate(details.Rental.DateFrom),
		DateTo:    formatDate(details.Rental.DateTo),
		Car: carInfo{
			CarUID:             details.Car.UID.String(),
			Brand:              details.Car.Brand,
			Model:              details.Car.Model,
			RegistrationNumber: details.Car.RegistrationNumber,
		},
		Payment: toPaymentInfo(details.Payment),
	}
}

func toRentalResponses(details []domain.RentalDetails) []rentalResponse {
	responses := make([]rentalResponse, 0, len(details))
	for _, d := range details {
		responses = append(responses, toRentalResponse(d))
	}
	return responses
}

func toCreateRentalResponse(booking domain.Booking) createRentalResponse {
	return createRentalResponse{
		RentalUID: booking.Rental.UID.String(),
		Status:    booking.Rental.Status,
		CarUID:    booking.Rental.CarUID.String(),
		DateFrom:  formatDate(booking.Rental.DateFrom),
		DateTo:    formatDate(booking.Rental.DateTo),
		Payment:   toPaymentInfo(booking.Payment),
	}
}
