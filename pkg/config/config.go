// Package config holds shared credentials, environment selection, and HTTP/error handling for
// the Payway (Decidir Argentina) Go SDK's resource clients.
//
// Unlike Talo (github.com/escapingnetwork/talo-go), which exchanges a client id/secret for a
// short-lived JWT, Decidir/Payway authenticates every request with a single static "apikey"
// header — there are two key tiers (public: tokenization only; private: payments/refunds/
// queries), but no token exchange, expiry, or refresh. A Config therefore holds exactly one key
// and is meant to be constructed once per key tier the caller needs (a server typically only
// ever needs a private-key Config; a public-key Config exists solely so pkg/tokenize can be
// tested end-to-end, since production tokenization happens client-side against Decidir directly,
// not through this SDK).
package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// Environment selects the Payway/Decidir API host.
type Environment string

const (
	Production Environment = "production"
	Sandbox    Environment = "sandbox"
)

// defaultBaseURLs: Payway's own hosts. The legacy Decidir sandbox host
// (developers.decidir.com) shares one backend with these — dropped 2026-09-03.
// Payway is also standing up an APIM gateway (api-sandbox.payway.com.ar,
// different error envelope) that its refund reference already points at; when
// the certification/production credentials are confirmed to belong to that
// stack, change these and cut a release. Callers can override today via
// config.WithBaseURL.
var defaultBaseURLs = map[Environment]string{
	Production: "https://ventasonline.payway.com.ar/api/v2",
	Sandbox:    "https://developers-ventasonline.payway.com.ar/api/v2",
}

// Config holds one API key (public or private, per the caller's choice) and the resolved base
// URL for either environment.
type Config struct {
	APIKey      string
	Environment Environment
	BaseURL     string // overrides Environment's default when set
	HTTPClient  *http.Client

	httpClient *http.Client
	baseURL    string
}

// Option is a functional option for configuring Config, mirroring talo-go's pattern.
type Option func(*Config)

// WithEnvironment sets the environment (Production or Sandbox). Defaults to Production.
func WithEnvironment(env Environment) Option {
	return func(c *Config) { c.Environment = env }
}

// WithBaseURL overrides the base URL completely, taking precedence over Environment.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) { c.BaseURL = baseURL }
}

// WithHTTPClient sets a custom *http.Client (e.g. for tests, via httptest.Server).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Config) { c.HTTPClient = hc }
}

// New creates a Config for a single API key. Pass the private key for payment/refund clients, or
// the public key for pkg/tokenize.
func New(apiKey string, opts ...Option) (*Config, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("payway: apiKey is required")
	}

	cfg := &Config{APIKey: apiKey, Environment: Production}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: defaultTimeout}
	}
	cfg.httpClient = cfg.HTTPClient

	base := cfg.BaseURL
	if base == "" {
		base = defaultBaseURLs[cfg.Environment]
		if base == "" {
			base = defaultBaseURLs[Production]
		}
	}
	cfg.baseURL = strings.TrimRight(base, "/")

	return cfg, nil
}

// DoRequest performs an authenticated API call and decodes a 2xx (or explicitly allowed) JSON
// response into out.
//
// Decidir's documented example payment response (see pkg/payment/response.go) is a flat JSON
// object with no {"data": ...} envelope, unlike Talo's API — so this does not attempt Talo's
// envelope-then-fallback unwrap. If a future endpoint turns out to be enveloped, prefer adding a
// narrow per-endpoint unwrap in that resource's client rather than reintroducing an unconditional
// guess here.
//
// extraOKStatuses lets a caller decode a non-2xx status as a normal body instead of an error —
// CONFIRMED needed for POST /payments, which returns HTTP 402 (not 200) for a business-rejected
// charge, with a fully valid Payment body attached (real id, status: "rejected", etc., verified
// directly against sandbox). Pass extra statuses only where confirmed true for that endpoint —
// this is not a general "ignore errors" escape hatch.
func (c *Config) DoRequest(ctx context.Context, method, path string, body, out any, extraOKStatuses ...int) error {
	fullURL := c.buildURL(path)

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("payway: marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)

	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !ok {
		for _, s := range extraOKStatuses {
			if resp.StatusCode == s {
				ok = true
				break
			}
		}
	}
	if !ok {
		return parseAPIError(resp.StatusCode, rawBody)
	}

	if out == nil || len(rawBody) == 0 {
		return nil
	}
	return json.Unmarshal(rawBody, out)
}

func (c *Config) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return c.baseURL + "/" + strings.TrimLeft(path, "/")
}

// decidirValidationError is Decidir's 400 body shape — CONFIRMED via direct sandbox testing
// against POST /tokens and POST /payments (both returned exactly
// {"error_type": "invalid_request_error", "validation_errors": [{"code": "...", "param": "..."}]}
// for missing/invalid fields). The bare-array fallback below is kept for defensiveness on
// endpoints not yet tested this way, but the object-wrapped shape is the confirmed one.
type decidirValidationError struct {
	Code  string `json:"code"`
	Param string `json:"param"`
}

func parseAPIError(statusCode int, rawBody []byte) error {
	var obj struct {
		ErrorType        string          `json:"error_type,omitempty"`
		ValidationErrors json.RawMessage `json:"validation_errors,omitempty"`
		Message          string          `json:"message,omitempty"`
		Errors           []struct {
			Status string `json:"status"`
			Code   string `json:"code"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		} `json:"errors,omitempty"`
	}
	_ = json.Unmarshal(rawBody, &obj)

	// validation_errors: array of {code,param}, OR an arbitrary-key object, OR absent.
	var vErrs []decidirValidationError
	var vMap map[string]string
	if len(obj.ValidationErrors) > 0 {
		if json.Unmarshal(obj.ValidationErrors, &vErrs) != nil {
			_ = json.Unmarshal(obj.ValidationErrors, &vMap)
		}
	}
	// Bare top-level array shape.
	if len(vErrs) == 0 && obj.ErrorType == "" && len(obj.Errors) == 0 {
		_ = json.Unmarshal(rawBody, &vErrs)
	}

	msg := obj.Message
	switch {
	case msg != "":
	case len(vErrs) > 0:
		msg = fmt.Sprintf("%s (%s)", vErrs[0].Code, vErrs[0].Param)
	case len(vMap) > 0:
		parts := make([]string, 0, len(vMap))
		for k, v := range vMap {
			parts = append(parts, k+"="+v)
		}
		sort.Strings(parts)
		msg = strings.Join(parts, ", ")
	case len(obj.Errors) > 0:
		e := obj.Errors[0]
		msg = strings.TrimSpace(e.Title + " " + e.Detail)
		if msg == "" {
			msg = e.Code
		}
	default:
		msg = fmt.Sprintf("HTTP %d", statusCode)
	}

	return &PaywayError{
		StatusCode:       statusCode,
		ErrorType:        obj.ErrorType,
		ValidationErrors: vErrs,
		Message:          msg,
		RawBody:          string(rawBody),
	}
}

// PaywayError represents an error returned by the Payway/Decidir API. Field population beyond
// StatusCode/RawBody depends on parseAPIError's shape guess above — treat ErrorType and
// ValidationErrors as best-effort until confirmed against sandbox.
type PaywayError struct {
	StatusCode       int                      `json:"status_code,omitempty"`
	ErrorType        string                   `json:"error_type,omitempty"`
	ValidationErrors []decidirValidationError `json:"validation_errors,omitempty"`
	Message          string                   `json:"message"`
	RawBody          string                   `json:"raw_body,omitempty"`
}

func (e *PaywayError) Error() string {
	return fmt.Sprintf("payway: %s (status=%d)", e.Message, e.StatusCode)
}
