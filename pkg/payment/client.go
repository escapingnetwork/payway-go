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
	// Create charges a card token. AmountCents on the request must be in cents, per Decidir's
	// documented "amount is an integer in cents" requirement.
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
	var out Payment
	if err := c.cfg.DoRequest(ctx, http.MethodPost, "/payments", req, &out); err != nil {
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
