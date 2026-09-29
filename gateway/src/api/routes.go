package api

import (
	"log/slog"
	"net/http"
	"time"
)

func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /manage/health", h.health)
	mux.HandleFunc("GET /api/v1/cars", h.listCars)
	mux.HandleFunc("GET /api/v1/rental", h.listRentals)
	mux.HandleFunc("POST /api/v1/rental", h.bookCar)
	mux.HandleFunc("GET /api/v1/rental/{rentalUid}", h.getRental)
	mux.HandleFunc("DELETE /api/v1/rental/{rentalUid}", h.cancelRental)
	mux.HandleFunc("POST /api/v1/rental/{rentalUid}/finish", h.finishRental)
	return logRequests(mux)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration", time.Since(started),
		)
	})
}
