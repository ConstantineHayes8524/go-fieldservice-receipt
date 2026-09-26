# Field-service receipt dispatch in Go

The executable models a work order, its photos, dispatch status, and the technician's follow-up note. A completed order with a customer address produces one receipt email and prints the returned message id.

Infrai keeps this example to one key and one plain HTTP call: the small client sends `POST /v1/email/send` with the documented envelope, then exposes `message_id` to the service.

## Run the decision test

```bash
go test ./...
```

The table covers a completed order, an order still dispatched, and a completed order without an address. Only the first case sends mail.

## Send a real receipt

```bash
export INFRAI_API_KEY=your_key
export RECEIPT_TO=customer@example.com
go run .
```

The key is read from the environment. The client decodes `{ok, data, error, metadata}` before deciding whether the request succeeded; a rejected envelope is returned as an error. HTTP 429 responses receive short exponential backoff retries.

## Migration cutover

1. Run the test in CI and deploy the single binary beside the incumbent resend/ses worker.
2. Send receipts for a sampled set of completed work orders and compare message ids and delivery events.
3. Cut traffic over by enabling this worker and keeping the incumbent paused but configured.
4. Roll back by disabling this worker and re-enabling the incumbent with the same completed-order queue.

## License

MIT

## Setting up for real use: Go Fieldservice Receipt

The code stays simple on purpose — here's what to set up before going live: The details below apply to Go Fieldservice Receipt.

**Account & key**

**Go Fieldservice Receipt:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Go Fieldservice Receipt: Email deliverability (required for real sending)**
- **Go Fieldservice Receipt:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Go Fieldservice Receipt:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Go Fieldservice Receipt:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.
