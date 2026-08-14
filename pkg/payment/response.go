package payment

// StatusDetails carries the finer-grained outcome of a charge. Ticket and CardAuthorizationCode
// are confirmed field names (from the same documented response example as Payment below); Error's
// shape when a payment is rejected was not shown in that example — kept as `any` until confirmed.
type StatusDetails struct {
	Ticket                string `json:"ticket,omitempty"`
	CardAuthorizationCode string `json:"card_authorization_code,omitempty"`
	Error                 any    `json:"error,omitempty"`
}

// Payment is Decidir's payment resource, returned by both Create and Get. Confirmed against the
// payway-ar/sdk-node-ventaonline README's documented example response (a flat JSON object, no
// envelope) — NOT independently verified against a live sandbox call.
//
// Status's only confirmed value from that example is "approved". Decidir is documented elsewhere
// (general product knowledge, NOT confirmed against this SDK's own sandbox testing) to also use
// "rejected", "pending", and "in_process" — treat the full vocabulary as unconfirmed until a real
// sandbox transaction exercises each path, and do not add a status-to-internal-state mapping
// inside this SDK; that belongs in the caller (matching how the Mobbex/MP adapters in
// prepa-backend keep status interpretation out of their raw API clients).
type Payment struct {
	ID                int64         `json:"id"`
	SiteTransactionID string        `json:"site_transaction_id"`
	PaymentMethodID   int           `json:"payment_method_id"`
	CardBrand         string        `json:"card_brand,omitempty"`
	AmountCents       int64         `json:"amount"`
	Currency          string        `json:"currency"`
	Status            string        `json:"status"`
	StatusDetails     StatusDetails `json:"status_details"`
	CustomerToken     string        `json:"customer_token,omitempty"`
}
