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

## Environments / base URLs

Payway's own hosts (the legacy Decidir sandbox host `developers.decidir.com` shared one backend
with these and was dropped 2026-09-03):

| Environment | Base URL |
|---|---|
| `config.Sandbox` | `https://developers-ventasonline.payway.com.ar/api/v2` |
| `config.Production` | `https://ventasonline.payway.com.ar/api/v2` |

Payway is also standing up an APIM gateway (`api-sandbox.payway.com.ar`, a different error
envelope — `parseAPIError` already tolerates it); the defaults will move there once the
credential stack is confirmed. Override either host today with `config.WithBaseURL`.

### Known sandbox test credentials (from official docs)

For a quick smoke test against `https://developers-ventasonline.payway.com.ar`, the official docs
publish these sandbox-only test keys for exercising the two-step (pre-authorization + capture) flow:

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
    Token:             cardToken, // obtained by the client tokenizing directly against Payway
    PaymentMethodID:   1,
    BinNumber:         "450799",
    AmountCents:        2550, // $25.50 — Payway amounts are integer cents (confirmed)
    Currency:           "ARS",
    Installments:        1,
})
```

### Distributed (marketplace) charge — split by amount

Set `PaymentType` to `payment.PaymentTypeDistributed` and pass one `SubPaymentRequest` per
participating Payway site; the legs must sum to the parent `AmountCents`.

```go
p, err := payClient.Create(ctx, payment.CreateRequest{
    SiteTransactionID: "order-123",
    Token:             cardToken,
    PaymentMethodID:   1,
    BinNumber:         "450799",
    AmountCents:        2550,
    Currency:          "ARS",
    Installments:      1,
    PaymentType:       payment.PaymentTypeDistributed,
    SubPayments: []payment.SubPaymentRequest{
        {SiteID: "04052018", Installments: 1, AmountCents: 2525}, // merchant
        {SiteID: "04052019", Installments: 1, AmountCents: 25},   // platform fee
    },
})
```

The response's `sub_payments[]` carries the per-leg outcome — `SubPayment.ID` (decoded from
`subpayment_id`), `SubPayment.AmountCents`, and `SubPayment.Status` (which can differ from the
parent `Payment.Status`). A distributed refund mirrors this: `refund.CreateRequest.SubPayments`
takes `[]payment.SubPaymentRequest` for a leg-by-leg reversal.

### Payment status vocabulary

Payway "Estado de las transacciones" (2026-09-03):

```
process    -> approved | group_rejected | group_annulled   (distributed validation)
approved   -> accredited (batch close) | annulled (reversal before close)
accredited -> refunded (full) | approved_with_refund (partial)
annulled   -> annulment_approved
refunded   -> refunded_approved
```

Synchronous charge outcomes a caller sees: `approved | rejected | review`. This SDK does **not**
map these to an internal state — that belongs in the caller (in prepa-backend,
`models.PaywayStatusToPaymentState`), matching the MP/Mobbex adapters.

## Status: many field shapes unverified

This SDK was scaffolded from documentation excerpts and a third-party SDK README, not a live
integration. Every place a shape is unconfirmed is marked in-code with a `NOT CONFIRMED` comment
explaining exactly what's uncertain and what to check in Decidir's sandbox before relying on it —
follow that discipline (see prepa-backend's `mobbex` adapter for the convention this mirrors)
rather than silently trusting or silently "fixing" those comments without testing against a real
sandbox call first.

### Confirmed against a real sandbox call (2026-08-14)

Using the known sandbox test credentials above and test card `4507990000004905`, the following
were verified directly against the Payway sandbox host (not just documentation):

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

### Confirmed against the official "Tablas de Referencia" (reference tables) page (2026-09-03)

- **`amount`'s unit is now settled: cents.** The docs' own field spec for `amount` reads "importe
  del pago" / "Importe minimo = 1 ($0.01)" — a minimum wire value of 1 mapping to one cent only
  makes sense if the field is integer cents throughout, resolving what was previously a genuine
  contradiction elsewhere in Decidir's docs. See `pkg/payment/response.go`'s `AmountCents` doc.
- **`payment_type`'s only two valid values are confirmed: `"single"` and `"distributed"`.**
- **Cybersource fraud detection is a per-merchant/site toggle, not universal** — "Si tu comercio lo
  tiene habilitado, es obligatorio enviar el objeto fraud_detection en flujos directos." This
  explains why the shared sandbox test key pair below couldn't complete a real approve/decline (see
  `pkg/payment/request.go`): that specific test site has it enabled. A real Prepa merchant sandbox
  account provisioned without it should charge cleanly.
- **98 payment method IDs** (`payment_method_id`, e.g. Visa=1, American Express=65, MasterCard
  Prisma=104, Diners Club=8) are documented on that page — not yet modeled as SDK constants (the
  full list is Argentina-card-network-specific and large; add them if/when a caller needs a
  human-readable→ID lookup rather than passing IDs through from elsewhere).
- HTTP status code table on that page (400 `malformed_request_error`, 401 `authentication_error`,
  402 `invalid_request_error`, 404 `not_found_error`, 409 `api_error`) **conflicts with what was
  independently observed against the live sandbox** (400 actually returned `invalid_request_error`,
  and 402 is used for a fully-formed *rejected payment*, not a request error) — the live sandbox
  behavior already encoded in `pkg/config/config.go` and `payment/client.go` is the one to trust;
  this doc table looks stale or describes a different product surface.

Also confirmed from `sdk-node-ventaonline`'s README and the official "Alcance" docs:
- `payment.Payment`'s enriched response fields (`date`, `tid`, `bin`, `installments`,
  `payment_type`, `sub_payments`, `site_id`, `status_details.address_validation_code`).
- **`amount` is integer centavos** — settled via the "Tablas de Referencia" field spec (min wire
  value `1` = `$0.01`); `prepa-backend` already sends cents. Do not change amount handling.
- **Refund is `POST /payments/{id}/refunds` and needs no `site_id`** — every refund-family method
  (`refund`, `partialRefund`, `deleteRefund`, `deletePartialRefund`) authenticates with the
  private apikey + payment id alone.
- Full payment `status` vocabulary — see "Payment status vocabulary" above.

NOT confirmed — verify against sandbox before shipping:
- Production host — the official docs give evidence for `live.decidir.com` for a *different*
  product (hosted checkout), not the REST Payments API this SDK calls, so
  `ventasonline.payway.com.ar` remains the default pending confirmation against a real
  production credential.
- `pkg/refund`'s response body shape (the endpoint path and no-`site_id` auth fact above are
  confirmed).
- A genuine end-to-end approve/decline charge — still blocked on getting a real (non-Cybersource-
  locked) Prepa merchant sandbox account; everything above was confirmed via request/response
  *shape* probing and documentation, not a completed charge.

## Testing

`go test ./...` — every resource client has `httptest.Server`-backed fixture tests from day one
(a gap `talo-go` left for later; don't leave it here too).
