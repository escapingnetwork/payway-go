package payment

// StatusDetails carries the finer-grained outcome of a charge. Ticket, CardAuthorizationCode, and
// AddressValidationCode are confirmed field names — from the official Decidir/Payway "Alcance"
// (SDK scope) documentation's payment-response examples, a stronger source than the
// sdk-node-ventaonline README this package was originally scaffolded from. Error's shape when a
// payment is rejected was not shown in any confirmed example — kept as `any` until confirmed.
type StatusDetails struct {
	Ticket                string `json:"ticket,omitempty"`
	CardAuthorizationCode string `json:"card_authorization_code,omitempty"`
	AddressValidationCode string `json:"address_validation_code,omitempty"`
	Error                 any    `json:"error,omitempty"`
}

// SubPayment is one entry of a distributed (marketplace/aggregator) payment's sub_payments array,
// confirmed from the official docs' "Pagos distribuidos" example. This SDK/integration does not
// build distributed payments (see CreateRequest's SubPayments doc), but a distributed payment's
// *response* could still come back with a populated array, so the shape is modeled for decoding
// completeness, not for construction.
type SubPayment struct {
	ID                    int64  `json:"id,omitempty"`
	SiteID                string `json:"site_id,omitempty"`
	Installments          int    `json:"installments,omitempty"`
	AmountCents           int64  `json:"amount,omitempty"`
	Ticket                string `json:"ticket,omitempty"`
	CardAuthorizationCode string `json:"card_authorization_code,omitempty"`
	TID                   string `json:"tid,omitempty"`
}

// Payment is Decidir's payment resource, returned by both Create and Get. Confirmed against the
// official Decidir/Payway "Alcance" (SDK scope) documentation's payment-response examples (a flat
// JSON object, no envelope) — this supersedes the sdk-node-ventaonline README this package was
// originally scaffolded from, and is NOT yet independently verified against a live sandbox call.
//
// Status's confirmed values from those examples are "approved" and "rejected"; Decidir is
// documented elsewhere (general product knowledge, NOT confirmed against this SDK's own sandbox
// testing) to also use "pending" and "in_process". Treat the full vocabulary as unconfirmed until
// a real sandbox transaction exercises each path, and do not add a status-to-internal-state
// mapping inside this SDK; that belongs in the caller (matching how the Mobbex/MP adapters in
// prepa-backend keep status interpretation out of their raw API clients).
//
// TID (Transaction ID) is confirmed present at the root level for Visa/Mastercard/Amex simple
// (non-distributed) payments; for distributed payments it appears per-SubPayment instead (see
// SubPayment.TID) rather than at the root.
//
// AmountCents' unit is CONFIRMED as cents (2026-09-03), settling what was previously a genuine
// contradiction in Decidir's own docs (a "Pagos distribuidos" clarification said cents; a "Pago
// Simple" example showed a decimal `amount: 25.50` with a "campo double" note). The official
// "Tablas de Referencia" page's field spec for `amount` states plainly: "importe del pago" /
// "Importe minimo = 1 ($0.01)" — a minimum wire value of 1 mapping to $0.01 is only consistent
// with cents. Also confirmed on the same page: the field's documented maximum is
// 922337203685 ($9223372036.85, i.e. int64-range cents), matching this field's int64 type.
type Payment struct {
	ID                         int64         `json:"id"`
	SiteTransactionID          string        `json:"site_transaction_id"`
	PaymentMethodID            int           `json:"payment_method_id"`
	CardBrand                  string        `json:"card_brand,omitempty"`
	AmountCents                int64         `json:"amount"`
	Currency                   string        `json:"currency"`
	Status                     string        `json:"status"`
	StatusDetails              StatusDetails `json:"status_details"`
	CustomerToken              string        `json:"customer_token,omitempty"`
	Date                       string        `json:"date,omitempty"`
	TID                        string        `json:"tid,omitempty"`
	BinNumber                  string        `json:"bin,omitempty"`
	Installments               int           `json:"installments,omitempty"`
	FirstInstallmentExpiration string        `json:"first_installment_expiration_date,omitempty"`
	PaymentType                string        `json:"payment_type,omitempty"`
	SubPayments                []SubPayment  `json:"sub_payments,omitempty"`
	SiteID                     string        `json:"site_id,omitempty"`
	Confirmed                  bool          `json:"confirmed,omitempty"`
}
