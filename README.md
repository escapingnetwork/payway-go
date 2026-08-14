# payway-go

Hand-rolled Go SDK for **Payway / Decidir** (Argentina), structured similarly to
[github.com/escapingnetwork/talo-go](https://github.com/escapingnetwork/talo-go) — one package
per resource, a shared `pkg/config` for HTTP/error handling, functional options for construction.

Ported from [payway-ar/sdk-node-ventaonline](https://github.com/payway-ar/sdk-node-ventaonline)'s
documented request/response shapes; cross-checked against
[payway-ar/sdk-javascript-ventaonline](https://github.com/payway-ar/sdk-javascript-ventaonline)
for the tokenization flow; further cross-checked against the official Decidir/Payway "Alcance"
(SDK scope) documentation, which is the strongest source used so far and superseded several of the
above where they conflicted (see field-level comments).

### Known sandbox test credentials (from official docs)

For a quick smoke test against `https://developers.decidir.com`, the official docs publish these
sandbox-only test keys for exercising the two-step (pre-authorization + capture) flow:

- Public test key: `41cbc74acc604a109157bb8394561d27`
- Private test key: `1fb6dc55c0a1489db411a8ee8f9c9707`

Sandbox-only — never use these (or any hardcoded key) against the production host.

## Two key tiers — do not mix them up

Decidir issues a **public** key and a **private** key per merchant.

- **Public key** — safe to embed client-side. Used only for `pkg/tokenize` (tokenizing raw card
  data). In production, mobile/web clients should call Decidir's tokenization endpoint directly
  rather than routing through a backend that holds this SDK — `pkg/tokenize` exists in this
  repo purely so the SDK's own test suite can exercise a full token→charge flow.
- **Private key** — server-side only, never sent to a client. Used for `pkg/payment` and
  `pkg/refund`.

Mixing these up either breaks tokenization (private key can't tokenize) or leaks a secret key to
a client (public key config used for a payment call will be rejected, but the intent is what
matters — never construct a `pkg/payment`/`pkg/refund` client with a public key, or vice versa).

## Usage

```go
cfg, err := config.New(privateAPIKey, config.WithEnvironment(config.Sandbox))
if err != nil {
    // ...
}

payClient := payment.NewClient(cfg)
p, err := payClient.Create(ctx, payment.CreateRequest{
    SiteTransactionID: "order-123",
    Token:             cardToken, // obtained by the client tokenizing directly against Decidir
    PaymentMethodID:   1,
    BinNumber:         "450799",
    AmountCents:        2550, // $25.50 — Decidir amounts are integer cents
    Currency:           "ARS",
    Installments:        1,
})
```

## Status: many field shapes unverified

This SDK was scaffolded from documentation excerpts and a third-party SDK README, not a live
integration. Every place a shape is unconfirmed is marked in-code with a `NOT CONFIRMED` comment
explaining exactly what's uncertain and what to check in Decidir's sandbox before relying on it —
follow that discipline (see prepa-backend's `mobbex` adapter for the convention this mirrors)
rather than silently trusting or silently "fixing" those comments without testing against a real
sandbox call first.

### Confirmed against a real sandbox call (2026-08-14)

Using the known sandbox test credentials above and test card `4507990000004905`, the following
were verified directly against `https://developers.decidir.com` (not just documentation):

- **`POST /tokens`'s real request/response shape** — completely different from this SDK's
  original guess. Request needs a nested `card_holder_identification: {type, number}` (not flat
  doc-type/number fields); response's id field is `id` (not `token`), and includes `bin`,
  `last_four_digits`, `expiration_month/year`, `cardholder` directly. See `pkg/tokenize`.
- **`POST /payments` requires `payment_type` and `sub_payments` on every request**, even a plain
  single-merchant charge — omitting either is a hard validation error. `Client.Create` now
  defaults `payment_type: "single"` and `sub_payments: []` so callers don't need to know this.
- **A business-rejected charge returns HTTP 402, not 200** — with a fully valid `Payment` body
  attached (real id, `status: "rejected"`, `status_details`, etc.). `Client.Create` decodes this
  normally now instead of surfacing it as a `PaywayError`; check `Payment.Status`, not `err`, to
  detect a decline.
- **The 400 error envelope is exactly** `{"error_type": "invalid_request_error",
  "validation_errors": [{"code": "...", "param": "..."}]}` — confirms `parseAPIError`'s primary
  shape guess.
- **Auth header `apikey`** and the sandbox host confirmed reachable and working end-to-end.

**Still NOT settled: `amount`'s unit (cents vs. major-units).** A real charge could not be pushed
through to a genuine approve/decline outcome — the published sandbox test key pair has full
Cybersource Decision Manager fraud detection enabled, which (after supplying `customer`,
`fraud_detection.bill_to`, `fraud_detection.purchase_totals`, etc.) ultimately demands a
`DeviceFingerprintID` — a browser-generated fraud-detection session id from a client-side JS
beacon, not obtainable from a server-side/curl test. `amount: 1` and `amount: 1000` both passed
every validation layer reached without an amount-specific rejection, which is weak evidence
against a strict pesos-scale minimum but does not prove the unit either way — see
`pkg/payment/response.go`'s `AmountCents` doc for the full reasoning. This SDK keeps the cents
interpretation; confirm against Prepa's actual (non-fraud-detection-locked) merchant sandbox
account, where a real approve/decline should be reachable without wrestling this specific test
key's Cybersource configuration.

Also confirmed from `sdk-node-ventaonline`'s README and the official "Alcance" docs:
- `payment.Payment`'s enriched response fields (`date`, `tid`, `bin`, `installments`,
  `payment_type`, `sub_payments`, `site_id`, `status_details.address_validation_code`).
- **Refund needs no `site_id`** — every refund-family SDK method (`refund`, `partialRefund`,
  `deleteRefund`, `deletePartialRefund`) takes only the private apikey + payment id.

NOT confirmed — verify against sandbox before shipping:
- Production host — see `pkg/config/config.go`'s expanded comment; the official docs give
  evidence for `live.decidir.com` for a *different* product (hosted checkout), not the REST
  Payments API this SDK calls, so `ventasonline.payway.com.ar` remains the default pending
  confirmation against a real production credential.
- Full payment `status` vocabulary beyond `"approved"`/`"rejected"`.
- `pkg/refund`'s response shape and exact endpoint path (only the request shape + no-site_id fact
  above are confirmed).

## Testing

`go test ./...` — every resource client has `httptest.Server`-backed fixture tests from day one
(a gap `talo-go` left for later; don't leave it here too).
