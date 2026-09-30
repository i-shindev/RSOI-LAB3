package domain

import "github.com/google/uuid"

type CarType string

const (
	CarTypeSedan    CarType = "SEDAN"
	CarTypeSUV      CarType = "SUV"
	CarTypeMinivan  CarType = "MINIVAN"
	CarTypeRoadster CarType = "ROADSTER"
)

type Car struct {
	UID                uuid.UUID
	Brand              string
	Model              string
	RegistrationNumber string
	Power              *int
	Price              int
	Type               *CarType
	Available          bool
}

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

type ListCarsInput struct {
	Page    *int
	Size    *int
	ShowAll bool
}

type ListCarsQuery struct {
	Page    int
	Size    int
	ShowAll bool
}

func (q ListCarsQuery) Offset() int {
	return (q.Page - 1) * q.Size
}

type CarsPage struct {
	Page          int
	TotalElements int
	Items         []Car
}
