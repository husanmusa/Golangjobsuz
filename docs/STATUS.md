# Current project state

This repository contains several prototypes and helpers for Golangjobsuz:

- **Safety controls HTTP service**: `/upload`, `/ai`, `/admin/audit`, `/metrics`, and `/healthz` endpoints with consent header enforcement, rate limiting, content-type filtering, structured logging, metrics, and alert hooks.
- **CLI recruiter toolkit**: `cmd/golangjobsuz` supports admin approvals/bans, recruiter search with filters (skills, location, seniority, recency), pagination, and profile views with redacted contacts and a contact-request CTA.
- **Draft/confirmation Telegram bot prototype (Python)**: `bot.py` stores drafts in `profiles.json`, lets users upload a document, review extracted fields, edit or re-parse, and confirm to create versioned profiles accessible via `/my_profile`.
- **Telegram ingestion bot (Go)**: `cmd/bot` validates uploads or URLs, downloads content, extracts text (PDF/DOCX with optional OCR), and writes raw/extracted data to local or S3 storage with configurable timeouts.
- **Job parsing API**: `cmd/server` parses job descriptions through an AI client, normalizes job fields, and can run in-memory or against Postgres with a baseline schema.
- **Broadcast/contact demo**: `cmd/app` formats job broadcasts with AI-style summaries, logs send attempts, supports dry runs with retries, and routes recruiter contact requests with logging.

# Next priorities

- Wire an AI parsing pipeline for resumes (OpenAI/Gemini) with structured JSON outputs and validation.
- Add bot UX for upload → AI parse → user confirmation, including retries and manual correction.
- Extend persistence for AI metadata (raw responses, model/version, extracted timestamps) and draft history.
- Strengthen governance: consent logging, stricter file scanning, rate limits, metrics on AI cost/latency, and audit logs for admin/recruiter actions.
