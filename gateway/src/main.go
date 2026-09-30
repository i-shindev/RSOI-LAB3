package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gateway/src/api"
	"gateway/src/clients"
	"gateway/src/config"
	"gateway/src/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	rq := service.NewRetryQueue(time.Second)
	gateway := service.NewGateway(
		clients.NewCarsClient(cfg.CarsServiceURL, cfg.RequestTimeout),
		clients.NewRentalClient(cfg.RentalServiceURL, cfg.RequestTimeout),
		clients.NewPaymentClient(cfg.PaymentServiceURL, cfg.RequestTimeout),
		rq,
	)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(api.NewHandlers(gateway)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("gateway listening", "addr", server.Addr,
			"cars", cfg.CarsServiceURL, "rental", cfg.RentalServiceURL, "payment", cfg.PaymentServiceURL)
		serverErr <- server.ListenAndServe()
	}()
	go rq.Run(ctx)

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
