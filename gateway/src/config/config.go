package config

import (
	"os"
	"time"
)

type Config struct {
	Port              string
	CarsServiceURL    string
	RentalServiceURL  string
	PaymentServiceURL string
	RequestTimeout    time.Duration
}

func Load() (Config, error) {
	timeout, err := time.ParseDuration(env("REQUEST_TIMEOUT", "5s"))
	if err != nil {
		return Config{}, err
	}
	return Config{
		Port:              env("PORT", "8080"),
		CarsServiceURL:    env("CARS_SERVICE_URL", "http://localhost:8070"),
		RentalServiceURL:  env("RENTAL_SERVICE_URL", "http://localhost:8060"),
		PaymentServiceURL: env("PAYMENT_SERVICE_URL", "http://localhost:8050"),
		RequestTimeout:    timeout,
	}, nil
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
