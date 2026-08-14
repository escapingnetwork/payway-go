// Package refund wraps Decidir's /payments/{id}/refunds resource. Construct with a PRIVATE-key
// config.Config, same as pkg/payment — only the account that collected a payment can refund it.
package refund

import (
	"context"
	"net/http"
	"net/url"

	"github.com/escapingnetwork/payway-go/pkg/config"
)

// Client is Decidir's refund resource.
type Client interface {
	// Create refunds paymentID. Pass req.AmountCents == nil for a full refund, or a value for a
	// partial refund.
	Create(ctx context.Context, paymentID string, req CreateRequest) (*Refund, error)
}

type client struct {
	cfg *config.Config
}

// NewClient builds a refund Client. cfg should hold Decidir's PRIVATE api key.
func NewClient(cfg *config.Config) Client {
	return &client{cfg: cfg}
}

// NOT CONFIRMED: the exact path (assumed POST /payments/{id}/refunds, mirroring the resource's
// own id-nesting convention used by payment.Client.Get) — the Node SDK README does not show the
// underlying REST path, only its own `sdk.refund(paymentId, cb)` wrapper. Verify against sandbox
// before relying on this in production.
func (c *client) Create(ctx context.Context, paymentID string, req CreateRequest) (*Refund, error) {
	var out Refund
	path := "/payments/" + url.PathEscape(paymentID) + "/refunds"
	if err := c.cfg.DoRequest(ctx, http.MethodPost, path, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
