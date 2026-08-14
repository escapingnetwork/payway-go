# payway-go

Hand-rolled Go SDK for **Payway / Decidir** (Argentina), structured similarly to
[github.com/escapingnetwork/talo-go](https://github.com/escapingnetwork/talo-go) — one package
per resource, a shared `pkg/config` for HTTP/error handling, functional options for construction.

Ported from [payway-ar/sdk-node-ventaonline](https://github.com/payway-ar/sdk-node-ventaonline)'s
documented request/response shapes; cross-checked against
[payway-ar/sdk-javascript-ventaonline](https://github.com/payway-ar/sdk-javascript-ventaonline)
for the tokenization flow.

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

Confirmed from `sdk-node-ventaonline`'s README:
- Auth header: `apikey` (single static key, no OAuth/token exchange).
- `payment.CreateRequest`/`payment.Payment` field names and the "amount is cents" rule.
- `refund.CreateRequest`'s partial-refund `amount` field and the `apikey` header usage.
- Sandbox host `https://developers.decidir.com` (this SDK appends `/api/v2`).

NOT confirmed — verify against sandbox before shipping:
- Production host (`ventasonline.payway.com.ar` vs `live.decidir.com` — see `pkg/config/config.go`).
- Full payment `status` vocabulary beyond `"approved"`.
- Error response shape (`pkg/config/config.go`'s `parseAPIError` probes a couple of documented-elsewhere
  shapes defensively, matching neither confirmed).
- `pkg/refund`'s response shape and exact endpoint path.
- `pkg/tokenize`'s entire request/response shape (the JS SDK's README only confirms the callback
  is `{token: "<uuid>"}` — nothing else).

## Testing

`go test ./...` — every resource client has `httptest.Server`-backed fixture tests from day one
(a gap `talo-go` left for later; don't leave it here too).
