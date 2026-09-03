package refund

import "github.com/escapingnetwork/payway-go/pkg/payment"

// CreateRequest asks Decidir to return money on an existing payment. AmountCents nil means a
// full refund (confirmed: the Node SDK's sdk.refund(paymentId, cb) takes no amount at all for a
// full refund — modeled here as an empty/omitted body rather than an explicit "full" flag, since
// the README shows no such flag). A non-nil AmountCents is a partial refund — confirmed field
// name "amount" (in cents, consistent with payment.CreateRequest.AmountCents) from the Node SDK's
// documented partialRefund example. NOT independently verified against a live sandbox call.
//
// SubPayments carries leg-by-leg amounts for reversing a distributed charge; omitted for a plain
// single-merchant refund.
type CreateRequest struct {
	AmountCents *int64                      `json:"amount,omitempty"`
	SubPayments []payment.SubPaymentRequest `json:"sub_payments,omitempty"`
}
