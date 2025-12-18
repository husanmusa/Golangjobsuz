# Golangjobsuz

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
