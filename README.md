# Golangjobsuz Safety Controls Service

This service demonstrates safety controls around file submissions and AI calls:

- Consent/privacy notice must be accepted via the `X-User-Consent: true` header before uploads are parsed, and each acceptance is logged with a request ID.
- Upload and AI endpoints are protected by an in-memory token-bucket rate limiter (5 requests/second, burst 5 per client IP).
- File uploads are scanned by content type and extension; disallowed formats are blocked and counted.
- Metrics capture submission volumes, AI success/failure, latency, and simulated AI cost; exposed at `/metrics`.
- Structured JSON logging includes request IDs and basic tracing; admin/recruiter actions are captured through `/admin/audit`.
- Error paths emit alerts through the notifier hook (ready to wire to email/Slack).

## Running

```bash
go run ./cmd/server
```

Endpoints:

- `POST /upload` — multipart form with `file`; requires `X-User-Consent: true` header.
- `POST /ai` — form field `prompt` to simulate an AI call.
- `POST /admin/audit` — form fields `actor`, `role`, and `action` for audit trails.
- `GET /metrics` — JSON metrics for scraping/logging.
- `GET /healthz` — simple health check.
