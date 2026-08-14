package tokenize

// Cardholder mirrors the tokenization response's echoed cardholder info.
type Cardholder struct {
	Identification CardHolderIdentification `json:"identification"`
	Name           string                   `json:"name"`
}

// Token is the tokenization result of POST /tokens. Every field below is CONFIRMED via a real
// sandbox call (2026-08-14) — a request against sandbox test card 4507990000004905 returned
// exactly this shape (HTTP 201):
//
//	{"id":"<uuid>","status":"active","card_number_length":16,"date_created":"...",
//	 "bin":"450799","last_four_digits":"4905","security_code_length":3,
//	 "expiration_month":12,"expiration_year":50,"date_due":"...",
//	 "cardholder":{"identification":{"type":"dni","number":"..."},"name":"..."}}
//
// This supersedes the SDK's original guess: the id field is "id", not "token" (the JS SDK
// callback's `{token: "<uuid>"}` shape, previously assumed to be this response verbatim, is
// evidently a wrapper/rename decidir.js applies client-side, not what the raw REST endpoint
// returns), and Bin/LastFourDigits/ExpirationMonth/ExpirationYear are all present directly on the
// token response — payment.CreateRequest's BinNumber can be sourced from here without a separate
// lookup.
type Token struct {
	ID               string     `json:"id"`
	Status           string     `json:"status,omitempty"`
	CardNumberLength int        `json:"card_number_length,omitempty"`
	DateCreated      string     `json:"date_created,omitempty"`
	BinNumber        string     `json:"bin,omitempty"`
	LastFourDigits   string     `json:"last_four_digits,omitempty"`
	SecurityCodeLen  int        `json:"security_code_length,omitempty"`
	ExpirationMonth  int        `json:"expiration_month,omitempty"`
	ExpirationYear   int        `json:"expiration_year,omitempty"`
	DateDue          string     `json:"date_due,omitempty"`
	Cardholder       Cardholder `json:"cardholder,omitempty"`
}
