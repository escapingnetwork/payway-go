package tokenize

// CreateRequest tokenizes raw card data. This package exists ONLY for this SDK's own test
// completeness (see package doc in client.go) — production mobile checkout calls Decidir's public
// tokenization endpoint directly from the app, not through this Go SDK, per the integration's
// confirmed design (see prepa-mobile's paywayTokenize.ts).
//
// Field names are NOT CONFIRMED: the payway-ar/sdk-node-ventaonline README's "internaltokens"
// example uses a differently-shaped, nested `card_data` object aimed at a distinct
// (establishment-scoped) tokenization variant, while the public decidir.js frontend flow
// (payway-ar/sdk-javascript-ventaonline) is documented only as posting a `data-decidir`-annotated
// HTML form with no JSON schema shown in its README. The flat field names below are this SDK's
// best guess, modeled on Decidir's commonly-documented public tokenization fields — verify every
// field name against a live sandbox response (the callback shape confirmed from the JS SDK's
// README is only `{token: "<uuid>"}`) before trusting this for anything beyond a starting point.
type CreateRequest struct {
	CardNumber          string `json:"card_number"`
	CardExpirationMonth string `json:"card_expiration_month"`
	CardExpirationYear  string `json:"card_expiration_year"`
	SecurityCode        string `json:"security_code"`
	CardHolderName      string `json:"card_holder_name"`
	CardHolderDocType   string `json:"card_holder_doc_type,omitempty"`
	CardHolderDocNumber string `json:"card_holder_doc_number,omitempty"`
}
