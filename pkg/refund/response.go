package refund

// Refund is Decidir's refund resource. NOT independently confirmed against a live sandbox
// response — the payway-ar Node SDK README documents the refund *request* shape but not its
// response body. Fields below are a reasonable guess mirroring payment.Payment's shape; verify
// and correct against sandbox before relying on any field other than a 2xx/4xx status check.
//
// CONFIRMED separately (from the official Decidir/Payway "Alcance" docs' SDK method signatures —
// sdk.refund(args, id, cb), sdk.partialRefund(args, id, cb), sdk.deleteRefund(args, id, cb),
// sdk.deletePartialRefund(args, id, cb)): every refund-family call is authenticated with only the
// apikey header plus the target payment's id — none of them take or need a site_id. This resolves
// the open question from the original implementation plan about whether SellerTokenRefunder's
// single-credential (private key only, no site_id) signature was sufficient for Payway — it is.
type Refund struct {
	ID          int64  `json:"id"`
	AmountCents int64  `json:"amount"`
	Status      string `json:"status,omitempty"`
}
