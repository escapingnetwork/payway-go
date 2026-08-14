// Package tokenize wraps Decidir's public card-tokenization endpoint. It exists solely so this
// SDK's own test suite can exercise a full token-then-charge flow against httptest fixtures
// without depending on decidir.js. It is NOT the production tokenization path: mobile checkout
// tokenizes cards by calling Decidir directly from the app (see prepa-mobile's
// paywayTokenize.ts), using the business's PUBLIC key, which is safe to embed client-side — that
// is the entire point of tokenization (PCI SAQ-A-EP scope reduction). Construct this package's
// client with a PUBLIC-key config.Config only; never load a private key into it.
package tokenize

import (
	"context"
	"net/http"

	"github.com/escapingnetwork/payway-go/pkg/config"
)

// Client is Decidir's public tokenization resource.
type Client interface {
	Create(ctx context.Context, req CreateRequest) (*Token, error)
}

type client struct {
	cfg *config.Config
}

// NewClient builds a tokenize Client. cfg should hold Decidir's PUBLIC api key — never the
// private key.
func NewClient(cfg *config.Config) Client {
	return &client{cfg: cfg}
}

// NOT CONFIRMED: the exact path (assumed POST /tokens, Decidir's conventional resource-root
// naming used elsewhere in this SDK) — verify against sandbox before relying on this.
func (c *client) Create(ctx context.Context, req CreateRequest) (*Token, error) {
	var out Token
	if err := c.cfg.DoRequest(ctx, http.MethodPost, "/tokens", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
