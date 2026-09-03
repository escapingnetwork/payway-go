// Package payment wraps Decidir's /payments resource: charging a card token and looking up a
// charge by id. Construct with a PRIVATE-key config.Config — payment creation is a server-side
// operation, never safe to call with the public tokenization key.
package payment

import (
	"context"
	"net/http"
	"net/url"

	"github.com/escapingnetwork/payway-go/pkg/config"
)

// Client is Decidir's payment resource.
type Client interface {
	// Create charges a card token. AmountCents on the request should be in cents, per Decidir's
	// documented "amount is an integer in cents" requirement — though this specific unit remains
	// genuinely unconfirmed, see CreateRequest's doc comment. A rejected charge (Payment.Status ==
	// "rejected") is returned as a normal *Payment, not an error — check Status, don't rely on err
	// to detect a decline (see the confirmed-402 note below).
	Create(ctx context.Context, req CreateRequest) (*Payment, error)
	// Get fetches a payment by its Decidir-assigned id (Payment.ID, stringified).
	Get(ctx context.Context, paymentID string) (*Payment, error)
}

type client struct {
	cfg *config.Config
}

// NewClient builds a payment Client. cfg should hold Decidir's PRIVATE api key.
func NewClient(cfg *config.Config) Client {
	return &client{cfg: cfg}
}

func (c *client) Create(ctx context.Context, req CreateRequest) (*Payment, error) {
	// PaymentType and SubPayments are hard-required by Decidir (confirmed, see CreateRequest's
	// doc) even for a plain non-aggregator charge; default them here so callers that only care
	// about single-merchant charges don't have to know about this requirement.
	if req.PaymentType == "" {
		req.PaymentType = PaymentTypeSingle
	}
	if req.SubPayments == nil {
		req.SubPayments = []SubPaymentRequest{}
	}

	var out Payment
	// CONFIRMED via direct sandbox testing: Decidir returns HTTP 402 (not 200) for a
	// business-rejected charge, with a fully valid Payment body (real id, status: "rejected",
	// status_details, etc.) — 402 there encodes a business outcome, not a transport/auth error, so
	// it must be decoded like any other success rather than turned into a PaywayError.
	if err := c.cfg.DoRequest(ctx, http.MethodPost, "/payments", req, &out, http.StatusPaymentRequired); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *client) Get(ctx context.Context, paymentID string) (*Payment, error) {
	var out Payment
	path := "/payments/" + url.PathEscape(paymentID)
	if err := c.cfg.DoRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
