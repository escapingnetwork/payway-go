package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoRequest_SetsAPIKeyHeaderAndDecodesFlatJSON(t *testing.T) {
	var gotAPIKey, gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("apikey")
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":123,"status":"approved"}`))
	}))
	defer srv.Close()

	cfg, err := New("test-private-key", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var out struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	if err := cfg.DoRequest(context.Background(), http.MethodPost, "/payments", map[string]string{"token": "tok"}, &out); err != nil {
		t.Fatalf("DoRequest: %v", err)
	}

	if gotAPIKey != "test-private-key" {
		t.Errorf("apikey header = %q, want %q", gotAPIKey, "test-private-key")
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/payments" {
		t.Errorf("path = %q, want /payments", gotPath)
	}
	if out.ID != 123 || out.Status != "approved" {
		t.Errorf("decoded out = %+v, want id=123 status=approved", out)
	}
}

func TestDoRequest_NonSuccessStatusReturnsPaywayError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		body, _ := json.Marshal(map[string]any{
			"validation_errors": []map[string]string{{"code": "invalid_param", "param": "token"}},
		})
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	cfg, err := New("test-key", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	err = cfg.DoRequest(context.Background(), http.MethodPost, "/payments", map[string]string{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	pwErr, ok := err.(*PaywayError)
	if !ok {
		t.Fatalf("error type = %T, want *PaywayError", err)
	}
	if pwErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", pwErr.StatusCode)
	}
	if len(pwErr.ValidationErrors) != 1 || pwErr.ValidationErrors[0].Code != "invalid_param" {
		t.Errorf("ValidationErrors = %+v, want one invalid_param entry", pwErr.ValidationErrors)
	}
}

func TestNew_RequiresAPIKey(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("expected error for empty apiKey, got nil")
	}
}

func TestEnvironmentDefaultBaseURLs(t *testing.T) {
	sandboxCfg, err := New("k", WithEnvironment(Sandbox))
	if err != nil {
		t.Fatal(err)
	}
	if got := sandboxCfg.buildURL("/x"); got != "https://developers-ventasonline.payway.com.ar/api/v2/x" {
		t.Errorf("sandbox buildURL = %q", got)
	}
	prodCfg, err := New("k", WithEnvironment(Production))
	if err != nil {
		t.Fatal(err)
	}
	if got := prodCfg.buildURL("/x"); got != "https://ventasonline.payway.com.ar/api/v2/x" {
		t.Errorf("production buildURL = %q", got)
	}
}
