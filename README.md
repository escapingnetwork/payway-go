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

Confirmed from `sdk-node-ventaonline`'s README and the official "Alcance" docs:
- Auth header: `apikey` (single static key, no OAuth/token exchange).
- `payment.CreateRequest`/`payment.Payment` field names, including several enriched response
  fields (`date`, `tid`, `bin`, `installments`, `payment_type`, `sub_payments`, `site_id`,
  `status_details.address_validation_code`) added after cross-checking the official docs' real
  response examples.
- `refund.CreateRequest`'s partial-refund `amount` field and the `apikey` header usage.
- **Refund needs no `site_id`** — every refund-family SDK method (`refund`, `partialRefund`,
  `deleteRefund`, `deletePartialRefund`) takes only the private apikey + payment id.
- Sandbox host `https://developers.decidir.com` (this SDK appends `/api/v2`).

NOT confirmed — verify against sandbox before shipping:
- Production host — see `pkg/config/config.go`'s expanded comment; the official docs give
  evidence for `live.decidir.com` for a *different* product (hosted checkout), not the REST
  Payments API this SDK calls, so `ventasonline.payway.com.ar` remains the default pending
  confirmation against a real production credential.
- Full payment `status` vocabulary beyond `"approved"`/`"rejected"`.
- **`amount`'s unit is genuinely contradictory in Decidir's own docs** — see the "Aclaración" vs.
  "Pago Simple" note in `pkg/payment/response.go`. This SDK keeps cents; confirm against a real
  sandbox charge before trusting it for real money.
- Error response shape (`pkg/config/config.go`'s `parseAPIError` probes a couple of documented-elsewhere
  shapes defensively, matching neither confirmed).
- `pkg/refund`'s response shape and exact endpoint path (only the request shape + no-site_id fact
  above are confirmed).
- `pkg/tokenize`'s entire request/response shape for the public-key client-side flow this package
  models. The official docs do give a concrete real shape, but for a differently-named endpoint
  (`internaltokens`, "tokenización interna") that may be a different product — see
  `pkg/tokenize/request.go` for the full reasoning and the concrete alternate shape to switch to
  if sandbox testing confirms it's the right one.

## Testing

`go test ./...` — every resource client has `httptest.Server`-backed fixture tests from day one
(a gap `talo-go` left for later; don't leave it here too).
