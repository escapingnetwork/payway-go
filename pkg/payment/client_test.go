package payment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/escapingnetwork/payway-go/pkg/config"
)

func TestClient_Create(t *testing.T) {
	var gotBody CreateRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/payments" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": 971344,
			"site_transaction_id": "id_1234567890",
			"payment_method_id": 65,
			"card_brand": "Visa",
			"amount": 2550,
			"currency": "ars",
			"status": "approved",
			"status_details": {"ticket": "4", "card_authorization_code": "203430"},
			"customer_token": "abc123"
		}`))
	}))
	defer srv.Close()

	cfg, err := config.New("private-key", config.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}
	client := NewClient(cfg)

	req := CreateRequest{
		SiteTransactionID: "id_1234567890",
		Token:             "tok-abc",
		PaymentMethodID:   1,
		BinNumber:         "450799",
		AmountCents:       2550,
		Currency:          "ARS",
		Installments:      1,
		PaymentType:       "single",
	}
	got, err := client.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.ID != 971344 {
		t.Errorf("ID = %d, want 971344", got.ID)
	}
	if got.Status != "approved" {
		t.Errorf("Status = %q, want approved", got.Status)
	}
	if got.AmountCents != 2550 {
		t.Errorf("AmountCents = %d, want 2550", got.AmountCents)
	}
	if got.StatusDetails.CardAuthorizationCode != "203430" {
		t.Errorf("StatusDetails.CardAuthorizationCode = %q", got.StatusDetails.CardAuthorizationCode)
	}
	if gotBody.AmountCents != 2550 {
		t.Errorf("request body amount = %d, want 2550", gotBody.AmountCents)
	}
}

func TestClient_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/payments/971344" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": 971344, "status": "approved", "amount": 2550, "currency": "ars"}`))
	}))
	defer srv.Close()

	cfg, err := config.New("private-key", config.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}
	client := NewClient(cfg)

	got, err := client.Get(context.Background(), "971344")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != 971344 || got.Status != "approved" {
		t.Errorf("got = %+v", got)
	}
}

func TestClient_Create_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"validation_errors":[{"code":"invalid_token","param":"token"}]}`))
	}))
	defer srv.Close()

	cfg, err := config.New("private-key", config.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}
	client := NewClient(cfg)

	_, err = client.Create(context.Background(), CreateRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
