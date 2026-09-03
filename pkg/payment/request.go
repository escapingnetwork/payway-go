package payment

// CreateRequest charges a previously generated card token, POSTed to /payments. Field names
// CONFIRMED via direct sandbox testing (2026-08-14, sandbox test card 4507990000004905):
//   - PaymentType and SubPayments are BOTH hard-required by Decidir even for a plain
//     single-merchant charge — omitting either returns
//     {"error_type":"invalid_request_error","validation_errors":[{"code":"param_required","param":"payment_type"},{"code":"param_required","param":"sub_payments"}]}.
//     "single" + an explicit empty array `[]` (not omitted, and not JSON null — Client.Create
//     defaults both when unset, see client.go) is CONFIRMED to be accepted for a non-aggregator
//     charge; this SDK/integration still does not build real distributed/aggregator payments, this
//     is purely satisfying a required-field validation.
//   - AmountCents' unit is CONFIRMED as cents (see payment.Payment's AmountCents doc for the
//     official docs' field-spec citation that settled this). A real approve/decline charge still
//     hasn't been completed against a live merchant account — the sandbox test credentials
//     published in Decidir's docs have Cybersource Decision Manager fraud detection enabled at the
//     site level, which demands a browser-generated DeviceFingerprintID this SDK's server-side
//     tests can't supply. The official docs describe Cybersource as a per-merchant/site toggle
//     ("Si tu comercio lo tiene habilitado, es obligatorio enviar el objeto fraud_detection..."),
//     not something universal — a real Prepa merchant sandbox account provisioned without it
//     should charge cleanly through this same request shape.
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
	PaymentType       string `json:"payment_type"`
	SubPayments       []any  `json:"sub_payments"`
}
