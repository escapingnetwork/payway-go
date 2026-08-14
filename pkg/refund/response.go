package refund

// Refund is Decidir's refund resource. NOT independently confirmed against a live sandbox
// response — the payway-ar Node SDK README documents the refund *request* shape but not its
// response body. Fields below are a reasonable guess mirroring payment.Payment's shape; verify
// and correct against sandbox before relying on any field other than a 2xx/4xx status check.
type Refund struct {
	ID          int64  `json:"id"`
	AmountCents int64  `json:"amount"`
	Status      string `json:"status,omitempty"`
}
