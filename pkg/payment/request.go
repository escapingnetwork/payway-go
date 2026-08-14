package payment

// CreateRequest charges a previously generated card token. Field names/shape confirmed against
// the payway-ar/sdk-node-ventaonline README's documented payment request example — NOT
// independently verified against a live sandbox call. In particular:
//   - AmountCents is documented explicitly as "an integer in cents (type long)" — e.g. 2550 for
//     $25.50 — do not pass a major-unit float here.
//   - PaymentType's only confirmed value is "single"; an "installments"-style value likely exists
//     for Installments > 1 but was not confirmed — verify against sandbox before assuming either
//     way, and prefer leaving PaymentType empty (Decidir may infer it from Installments) if the
//     sandbox rejects a guessed value.
//   - SubPayments is present so the wire shape matches Decidir's aggregator/marketplace field,
//     but is deliberately left unused (nil/empty) — this SDK/integration is scoped to
//     single-merchant, non-aggregator charges only. Do not populate it without revisiting that
//     scope decision first.
type CreateRequest struct {
	SiteTransactionID string `json:"site_transaction_id"`
	Token             string `json:"token"`
	UserID            string `json:"user_id,omitempty"`
	PaymentMethodID   int    `json:"payment_method_id"`
	BinNumber         string `json:"bin"`
	AmountCents       int64  `json:"amount"`
	Currency          string `json:"currency"`
	Installments      int    `json:"installments"`
	Description       string `json:"description,omitempty"`
	PaymentType       string `json:"payment_type,omitempty"`
	SubPayments       []any  `json:"sub_payments,omitempty"`
}
