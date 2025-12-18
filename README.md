# Golangjobsuz bot prototype

This repository contains a lightweight prototype for managing recruiter approvals and searching candidate profiles via CLI commands that mimic bot behaviors.

## Features
- **Admin approvals/bans**: `/admin --action approve|ban --user <id>` updates recruiter roles and stores status in both `users` and `recruiter_access` collections.
- **Search with filters & pagination**: `/search --skills golang,grpc --location Tashkent --seniority mid --days 14 --page 1 --page-size 5` filters by skills (substring match), location, seniority, and profile recency.
- **Profile view with redacted contacts**: `/profile --id p1` shows masked contact info plus a CTA to request contact through the bot.

## Running
```bash
# build
go build ./cmd/golangjobsuz

# examples
./golangjobsuz admin --action approve --user u3 --notes "verified" --admin admin
./golangjobsuz search --skills golang,grpc --location Tashkent --seniority mid --days 30 --page 1 --page-size 5
./golangjobsuz profile --id p1
```

Data is persisted to `data/store.json`; a default set of users and profiles is created on first run.
# Golangjobsuz bot

Telegram bot prototype for managing candidate profiles with a draft-to-confirmation review flow.

## Features
- Upload a document to create a `draft_profile`, trigger parsing, and show a processing message.
- Review extracted fields with inline keyboards to edit, re-parse, or manually correct data.
- Confirm drafts into persisted profiles with version history.
- Use `/my_profile` to view, update (create a new draft from the latest profile), or delete the latest saved profile.

## Running locally
1. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```
2. Set your bot token:
   ```bash
   export TELEGRAM_BOT_TOKEN="<your token>"
   ```
3. Start the bot:
   ```bash
   python bot.py
   ```

Profiles are stored in `profiles.json` in the repository root. Each confirmation appends a new version while keeping drafts separate.
# Golangjobsuz Telegram Ingestion Bot

This bot accepts documents or direct links from Telegram, validates them, downloads the content, extracts text (including an optional OCR fallback for images), and stores both the raw file and extracted text in either local storage or S3.

## Features
- Telegram handler for document uploads and URL submissions with MIME type and size validation.
- Downloads files directly through the Telegram API with clear user-facing retry guidance.
- Text extraction for PDF and DOCX files, plus optional OCR via an HTTP endpoint for images.
- Pluggable storage backends (local filesystem or S3) for raw and extracted text outputs.
- Operation timeouts to prevent long-running tasks from blocking the bot.

## Running the bot
1. Ensure Go 1.20+ is installed.
2. Export the required environment variables:
   - `TELEGRAM_BOT_TOKEN` (required)
   - `LOCAL_STORAGE_PATH` (optional, defaults to `data/`)
   - `S3_BUCKET` and optional `S3_PREFIX` to use S3 instead of local storage.
3. Build and run:
   ```bash
   go run ./cmd/bot
   ```

## OCR configuration
The built-in OCR hook posts the image bytes to an HTTP endpoint. Configure `extract.Extractor{OCR: &extract.HttpOCRProvider{URL: "https://your-ocr-endpoint", Timeout: 30 * time.Second}}` when constructing the service to enable OCR for images.

## Testing and development notes
Network access to download Go modules may be restricted in some environments. If `go mod tidy` or `go test ./...` fails with proxy errors, ensure module downloads are allowed or use a module proxy that is accessible from your environment.
# Golangjobsuz

This repository contains a lightweight toolkit for broadcasting job vacancies and routing recruiter contact requests.

## Features
- Format job broadcasts with an AI-style summary and key fields, then post them to a configured vacancies channel.
- Track broadcast attempts in a persistent JSON log with dry-run support and retry handling for transient send failures.
- Route recruiter contact requests directly to a job seeker or through an admin relay, with logging for each attempt.

## Quick start
Run the demo app to see both flows in action:

```bash
go run ./cmd/app
```

Broadcast records are written to `data/broadcasts.json` and contact request logs are kept in memory for the demo run.
Local development helpers for the Golangjobsuz bot and API.

## Configuration

Copy `.env.example` to `.env` and fill in the values for the bot token, webhook details, database DSN, channel/admin IDs, AI provider credentials, content limits, and runtime settings.

```bash
cp .env.example .env
```

### Key settings
- `BOT_TOKEN`: Bot authentication token.
- `WEBHOOK_URL` / `WEBHOOK_SECRET`: Endpoint and secret for inbound events.
- `DATABASE_DSN`: PostgreSQL connection string used by migrations and the app.
- `CHANNEL_ID` / `ADMIN_IDS`: Target channel and admin user IDs (comma separated).
- `AI_PROVIDER` / `AI_MODEL` and fallbacks: Provider and model names with `AI_API_KEY`.
- Limits and storage: `MAX_FILE_BYTES`, `MAX_FILE_TYPES`, `MAX_MESSAGE_BYTES`, `TEMP_STORAGE_PATH`.
- Timeouts: `REQUEST_TIMEOUT_SECONDS`, `RESPONSE_TIMEOUT_SECONDS`.

## Database migrations

Migrations use the `golang-migrate` CLI. Ensure `DATABASE_DSN` is exported in your shell or loaded from `.env`.

```bash
# Apply migrations
make migrate-up

# Roll back the most recent migration
make migrate-down
```

Tables include users, published and draft profiles (with summaries, links, and AI metadata), broadcasts, audit logs, and recruiter access controls.
