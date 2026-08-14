package tokenize

// CreateRequest tokenizes raw card data. This package exists ONLY for this SDK's own test
// completeness (see package doc in client.go) — production mobile checkout calls Decidir's public
// tokenization endpoint directly from the app, not through this Go SDK, per the integration's
// confirmed design (see prepa-mobile's paywayTokenize.ts).
//
// Field names are still NOT CONFIRMED for this specific (public-key, client-side) tokenization
// flow. The official Decidir/Payway "Alcance" docs DO now give a concrete, real JSON shape for a
// token request — but for `internaltokens()`, a distinctly-named ("tokenización interna")
// endpoint using a nested `card_data` object plus a sibling `establishment_number` field:
//
//	{
//	  "card_data": {
//	    "card_number": "...", "expiration_date": "MMYY" (single combined field, not separate
//	    month/year), "card_holder": "...", "security_code": "...",
//	    "account_number": "..." (long numeric string, purpose unclear — possibly a merchant
//	    sub-account/virtual-card reference, not customer-supplied; marked "obligatorio" in the
//	    spec but its real meaning needs confirming with Payway support before it can be filled in
//	    correctly), "email_holder": "..."
//	  },
//	  "establishment_number": "..." (plausibly the merchant's site_id, unconfirmed)
//	}
//
// Deliberately NOT adopted here: "interna" plus the establishment_number/account_number fields
// suggest this may be a different product (e.g. POS/omnichannel tokenization) rather than the
// standard e-commerce public-key client-side tokenization this package models — conflating the
// two without confirmation risks being wrong in a new way, not just an unconfirmed way. If sandbox
// testing confirms internaltokens() IS the right endpoint for this integration, replace the fields
// below with the card_data/establishment_number shape above; until then this keeps the original
// flat best-guess field names, modeled on Decidir's commonly-documented public tokenization
// fields. The callback shape confirmed from the JS SDK's README is only `{token: "<uuid>"}`.
type CreateRequest struct {
	CardNumber          string `json:"card_number"`
	CardExpirationMonth string `json:"card_expiration_month"`
	CardExpirationYear  string `json:"card_expiration_year"`
	SecurityCode        string `json:"security_code"`
	CardHolderName      string `json:"card_holder_name"`
	CardHolderDocType   string `json:"card_holder_doc_type,omitempty"`
	CardHolderDocNumber string `json:"card_holder_doc_number,omitempty"`
}
