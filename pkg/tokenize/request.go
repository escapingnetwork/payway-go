package tokenize

// CardHolderIdentification is the cardholder's national ID, required by POST /tokens — CONFIRMED
// via direct sandbox testing (2026-08-14): omitting it entirely returns
// {"error_type":"invalid_request_error","validation_errors":[{"code":"param_required","param":"card_holder_identification"}]}
// and a nested {type, number} object (not flat card_holder_doc_type/card_holder_doc_number
// fields, this SDK's original guess) is what the endpoint actually accepts. "dni" was confirmed as
// a valid Type value against the sandbox test card; other Argentine id types (e.g. "cuit") are
// plausible but not independently tested.
type CardHolderIdentification struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

// CreateRequest tokenizes raw card data via POST /tokens. This package exists ONLY for this SDK's
// own test completeness (see package doc in client.go) — production mobile checkout calls
// Decidir's public tokenization endpoint directly from the app, not through this Go SDK, per the
// integration's confirmed design (see prepa-mobile's paywayTokenize.ts).
//
// Field names CONFIRMED via a real sandbox call (2026-08-14, sandbox test card
// 4507990000004905): a request built from exactly these fields returned HTTP 201 with a valid
// token. This supersedes the SDK's original flat-guess shape (which had CardHolderDocType/
// CardHolderDocNumber instead of the nested CardHolderIdentification below) and the differently-
// shaped `internaltokens()` structure documented separately in Decidir's official "Alcance" docs —
// that one is a distinct, unconfirmed-relevance endpoint (see git history for the prior
// investigation); this file now reflects what was actually verified to work.
type CreateRequest struct {
	CardNumber               string                   `json:"card_number"`
	CardExpirationMonth      string                   `json:"card_expiration_month"`
	CardExpirationYear       string                   `json:"card_expiration_year"`
	SecurityCode             string                   `json:"security_code"`
	CardHolderName           string                   `json:"card_holder_name"`
	CardHolderIdentification CardHolderIdentification `json:"card_holder_identification"`
}
