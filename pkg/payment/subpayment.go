package payment

// SubPaymentRequest is one leg of a distributed (marketplace) charge or refund.
// By-percentage distribution sends an empty slice; by-amount sends one entry per
// participating Payway site, and the entries must sum to the parent amount.
type SubPaymentRequest struct {
	SiteID       string `json:"site_id"`
	Installments int    `json:"installments"`
	AmountCents  int64  `json:"amount"`
}
