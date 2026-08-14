package refund

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/escapingnetwork/payway-go/pkg/config"
)

func TestClient_Create_FullRefund(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/payments/971344/refunds" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": 55, "amount": 2550, "status": "refunded"}`))
	}))
	defer srv.Close()

	cfg, err := config.New("private-key", config.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}
	client := NewClient(cfg)

	got, err := client.Create(context.Background(), "971344", CreateRequest{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ID != 55 || got.Status != "refunded" {
		t.Errorf("got = %+v", got)
	}
	if _, present := gotBody["amount"]; present {
		t.Errorf("full refund body should omit amount, got %+v", gotBody)
	}
}

func TestClient_Create_PartialRefund(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": 55, "amount": 1050, "status": "refunded"}`))
	}))
	defer srv.Close()

	cfg, err := config.New("private-key", config.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}
	client := NewClient(cfg)

	amount := int64(1050)
	got, err := client.Create(context.Background(), "971344", CreateRequest{AmountCents: &amount})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.AmountCents != 1050 {
		t.Errorf("AmountCents = %d, want 1050", got.AmountCents)
	}
	if amt, _ := gotBody["amount"].(float64); int64(amt) != 1050 {
		t.Errorf("request body amount = %v, want 1050", gotBody["amount"])
	}
}
