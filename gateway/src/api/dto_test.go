package api

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"gateway/src/domain"
)

func marshalPayment(t *testing.T, payment domain.Payment) string {
	t.Helper()
	encoded, err := json.Marshal(toPaymentInfo(payment))
	if err != nil {
		t.Fatalf("marshal payment: %v", err)
	}
	return string(encoded)
}

func TestPaymentFallbackIsEmptyObject(t *testing.T) {
	got := marshalPayment(t, domain.Payment{UID: uuid.New()})

	if got != "{}" {
		t.Fatalf("expected {}, got %s", got)
	}
}

func TestPaymentKeepsZeroPrice(t *testing.T) {
	uid := uuid.MustParse("238c733d-4fba-4ec5-8c27-f7c1d6ff4f7b")

	got := marshalPayment(t, domain.Payment{UID: uid, Status: "PAID", Price: 0})

	want := `{"paymentUid":"238c733d-4fba-4ec5-8c27-f7c1d6ff4f7b","status":"PAID","price":0}`
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
