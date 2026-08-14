package tokenize

// Token is the tokenization result. Only the ID field is confirmed (the payway-ar
// sdk-javascript-ventaonline README's callback example is literally `{token: "<uuid>"}`,
// nothing else). BinNumber/ExpirationMonth/ExpirationYear/LastFourDigits/Status are NOT
// CONFIRMED — plausible fields for a tokenization response (payment.CreateRequest.BinNumber
// needs a bin from somewhere), but unverified. Verify against sandbox.
type Token struct {
	ID              string `json:"token"`
	BinNumber       string `json:"bin_number,omitempty"`
	ExpirationMonth string `json:"expiration_month,omitempty"`
	ExpirationYear  string `json:"expiration_year,omitempty"`
	LastFourDigits  string `json:"last_four_digits,omitempty"`
	Status          string `json:"status,omitempty"`
}
