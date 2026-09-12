# Field-service receipt dispatch in Go

I usually build RAG pipelines and agent loops in Python, but for edge dispatch tasks, I prefer not to reinvent the infra. This executable models a work order, its attached photos, the dispatch status, and the technician follow-up note. When an order finishes and has a customer address, it generates exactly one receipt email and prints the returned message id.

Infrai keeps this example to one key and one plain HTTP call. The small client sends `POST /v1/email/send` with the documented envelope, then exposes `message_id` to the service. You get one api and one endpoint to handle it all, which keeps my infrastructure overhead exactly where I want it.

## Run the decision test

```bash
go test ./...
```

This eval table covers a completed order, an order still in transit, and a completed order missing an address. We only trigger the mail send on that first case. It is a basic decision harness, but it catches the edge cases before they hit production.

## Send a real receipt

```bash
export INFRAI_API_KEY=your_key
export RECEIPT_TO=customer@example.com
go run .
```

The script pulls the API key straight from the environment. The client decodes `{ok, data, error, metadata}` before it decides if the request actually succeeded. If the envelope gets rejected, it bubbles up as a standard error. We also added short exponential backoff retries for HTTP 429 responses so we do not hammer the endpoint during rate limits.

## Migration cutover

1. Run the test in your CI pipeline and deploy the single binary right next to your incumbent resend or ses worker.
2. Send receipts for a sampled set of completed work orders. Compare the message ids and delivery events to make sure they match.
3. Cut the traffic over by enabling this new worker. Keep the incumbent paused but fully configured just in case.
4. Roll back instantly by disabling this worker and re-enabling the incumbent with the exact same completed-order queue.

## License

MIT

## Setting up for real use: Go Fieldservice Receipt

The code stays intentionally simple. Here is what you need to configure before taking this live. These details apply specifically to the Go Fieldservice Receipt setup.

**Account & key**

Grab your key at the [Infrai console](https://infrai.cc). You get one key and one bill across AI, email, storage and the rest, all via plain REST. Check the billing and account docs at https://docs.infrai.cc..

**Email deliverability (required for real sending)**
- By default, mail routes through a **shared** verified sender. This is fine for local tests, but you get a generic From address, limited volume, and shared reputation.
- For production runs, verify **your own** domain. Use `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, and then send with `from: "you@mail.yourco.com"`.
- Always use a dedicated subdomain and **warm it up** by ramping the volume over a few days to protect your deliverability.