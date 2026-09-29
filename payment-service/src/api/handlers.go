package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"payment-service/src/domain"
	"payment-service/src/service"
)

type Handlers struct {
	payments *service.PaymentService
}

func NewHandlers(payments *service.PaymentService) *Handlers {
	return &Handlers{payments: payments}
}

func (h *Handlers) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) createPayment(w http.ResponseWriter, r *http.Request) {
	var request createPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeValidationError(w, "Malformed JSON body", nil)
		return
	}
	payment, err := h.payments.Create(r.Context(), domain.NewPaymentInput{Price: request.Price})
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/payments/"+payment.UID.String())
	writeJSON(w, http.StatusCreated, toPaymentResponse(payment))
}

func (h *Handlers) getPayment(w http.ResponseWriter, r *http.Request) {
	uid, ok := paymentUID(w, r)
	if !ok {
		return
	}
	payment, err := h.payments.Get(r.Context(), uid)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPaymentResponse(payment))
}

func (h *Handlers) cancelPayment(w http.ResponseWriter, r *http.Request) {
	uid, ok := paymentUID(w, r)
	if !ok {
		return
	}
	if err := h.payments.Cancel(r.Context(), uid); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func paymentUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	uid, err := uuid.Parse(r.PathValue("paymentUid"))
	if err != nil {
		writeValidationError(w, "Validation failed", []domain.FieldError{
			{Field: "paymentUid", Message: "must be a valid UUID"},
		})
		return uuid.Nil, false
	}
	return uid, true
}
