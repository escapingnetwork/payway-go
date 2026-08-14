package tokenize

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/escapingnetwork/payway-go/pkg/config"
)

func TestClient_Create(t *testing.T) {
	var gotAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("apikey")
		if r.Method != http.MethodPost || r.URL.Path != "/tokens" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": "99ab0740-4ef9-4b38-bdf9-c4c963459b22", "status": "active", "bin": "450799", "last_four_digits": "4905"}`))
	}))
	defer srv.Close()

	cfg, err := config.New("public-key", config.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("config.New: %v", err)
	}
	client := NewClient(cfg)

	got, err := client.Create(context.Background(), CreateRequest{
		CardNumber:          "4507990000004905",
		CardExpirationMonth: "12",
		CardExpirationYear:  "50",
		SecurityCode:        "123",
		CardHolderName:      "Jorge Jorgelin",
		CardHolderIdentification: CardHolderIdentification{
			Type:   "dni",
			Number: "12345678",
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ID != "99ab0740-4ef9-4b38-bdf9-c4c963459b22" {
		t.Errorf("Token.ID = %q", got.ID)
	}
	if gotAPIKey != "public-key" {
		t.Errorf("apikey header = %q, want public-key", gotAPIKey)
	}
}
