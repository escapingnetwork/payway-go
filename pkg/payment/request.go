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
//   - AmountCents' UNIT REMAINS UNCONFIRMED. Decidir's own docs are internally contradictory (see
//     payment.Payment's AmountCents doc), and a real sandbox charge could not be pushed through to
//     a genuine approve/decline outcome to settle it empirically — the sandbox test credentials
//     published in Decidir's docs have full Cybersource Decision Manager fraud detection enabled,
//     which demands a browser-generated DeviceFingerprintID (a client-side JS beacon session, not
//     obtainable via a server-side/curl test) before evaluating the charge at all. amount=1 and
//     amount=1000 both passed every validation layer reached without an amount-specific error,
//     which is weak evidence against a strict minimum but does not prove cents vs. major-units.
//     Keeps the cents interpretation (matches the more explicit of Decidir's two contradictory doc
//     statements, and the Mobbex/MP convention elsewhere in prepa-backend) — confirm against a
//     charge on Prepa's actual (non-fraud-detection-locked) sandbox merchant account before
//     trusting this for real money.
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
