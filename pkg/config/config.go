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

// defaultBaseURLs per the resource list the SDK was scaffolded from: Payway's own production
// host (ventasonline.payway.com.ar) rather than Decidir's generic live.decidir.com, since this
// SDK targets the Payway (Argentina) product specifically. STILL NOT INDEPENDENTLY CONFIRMED for
// this REST payment API against a live call. The official Decidir/Payway "Alcance" docs' hosted
// checkout links (a different product — payment-link redirects, not what this SDK calls) do
// consistently use developers.decidir.com (sandbox) / live.decidir.com (production) with a
// /web/checkout/{payment_id} path, no /api/v2 segment — which is evidence for live.decidir.com as
// at least one valid production host, but doesn't settle whether a Payway-branded merchant's REST
// Payments API calls route through ventasonline.payway.com.ar or live.decidir.com/api/v2 (Payway
// is a Decidir white-label, so the underlying backend may accept either host, or may require the
// Payway-branded one specifically for these credentials). Verify against sandbox/production docs
// or a real production credential before relying on this for real money, and correct here if it
// differs.
var defaultBaseURLs = map[Environment]string{
	Production: "https://ventasonline.payway.com.ar/api/v2",
	Sandbox:    "https://developers.decidir.com/api/v2",
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

// DoRequest performs an authenticated API call and decodes a 2xx JSON response into out.
//
// Decidir's documented example payment response (see pkg/payment/response.go) is a flat JSON
// object with no {"data": ...} envelope, unlike Talo's API — so this does not attempt Talo's
// envelope-then-fallback unwrap. If a future endpoint turns out to be enveloped, prefer adding a
// narrow per-endpoint unwrap in that resource's client rather than reintroducing an unconditional
// guess here.
func (c *Config) DoRequest(ctx context.Context, method, path string, body, out any) error {
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
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

// decidirValidationError is ONE of several NOT CONFIRMED shapes Decidir's docs and various
// third-party integration reports describe for 4xx bodies — commonly a bare array of
// {code, param} objects, sometimes wrapped under "validation_errors, sometimes with a top-level
// "error_type". parseAPIError tries the array-first shape (the most commonly cited one) and
// falls back to a generic message extraction, the same "probe several shapes defensively" idea
// talo-go's parseAPIError uses for its own not-fully-consistent upstream. VERIFY THE REAL SHAPE
// AGAINST SANDBOX before depending on ErrorType/ValidationErrors for user-facing error copy.
type decidirValidationError struct {
	Code  string `json:"code"`
	Param string `json:"param"`
}

func parseAPIError(statusCode int, rawBody []byte) error {
	var withArray struct {
		ErrorType        string                   `json:"error_type,omitempty"`
		ValidationErrors []decidirValidationError `json:"validation_errors,omitempty"`
		Message          string                   `json:"message,omitempty"`
	}
	_ = json.Unmarshal(rawBody, &withArray)

	// Some Decidir error responses are documented as a bare top-level array of
	// {code, param} rather than wrapped in an object — try that shape too if the object
	// decode found nothing useful.
	var bareArray []decidirValidationError
	if len(withArray.ValidationErrors) == 0 && withArray.ErrorType == "" {
		_ = json.Unmarshal(rawBody, &bareArray)
	}

	validationErrors := withArray.ValidationErrors
	if len(validationErrors) == 0 {
		validationErrors = bareArray
	}

	msg := withArray.Message
	if msg == "" && len(validationErrors) > 0 {
		msg = fmt.Sprintf("%s (%s)", validationErrors[0].Code, validationErrors[0].Param)
	}
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", statusCode)
	}

	return &PaywayError{
		StatusCode:       statusCode,
		ErrorType:        withArray.ErrorType,
		ValidationErrors: validationErrors,
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
