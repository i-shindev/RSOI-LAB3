package api

import "cars-service/src/domain"

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

func toCarResponse(car domain.Car) carResponse {
	var carType *string
	if car.Type != nil {
		t := string(*car.Type)
		carType = &t
	}
	return carResponse{
		CarUID:             car.UID.String(),
		Brand:              car.Brand,
		Model:              car.Model,
		RegistrationNumber: car.RegistrationNumber,
		Power:              car.Power,
		Type:               carType,
		Price:              car.Price,
		Available:          car.Available,
	}
}

type paginationResponse struct {
	Page          int           `json:"page"`
	PageSize      int           `json:"pageSize"`
	TotalElements int           `json:"totalElements"`
	Items         []carResponse `json:"items"`
}

func toPaginationResponse(page domain.CarsPage) paginationResponse {
	items := make([]carResponse, 0, len(page.Items))
	for _, car := range page.Items {
		items = append(items, toCarResponse(car))
	}
	return paginationResponse{
		Page:          page.Page,
		PageSize:      len(items),
		TotalElements: page.TotalElements,
		Items:         items,
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
