# Golangjobsuz Runbook

## Environment variables
- `DATABASE_URL`: Postgres connection string. Defaults to in-memory storage; when provided in containerized deployments it can point at Postgres for persistence.
- `PORT`: HTTP listen port. Default `8080`.
- `AI_ENDPOINT`: HTTP endpoint used by the AI client for completion. Default `http://localhost:18081/complete`.

## Migrations
The repository automatically applies the baseline schema on startup via `InitSchema`. For manual execution run:

```bash
go run ./cmd/server # initializes schema on startup
```

Or apply manually:

```sql
CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    company TEXT NOT NULL,
    location TEXT NOT NULL,
    description TEXT NOT NULL
);
```

## Admin commands
- Seed a job record for testing: `psql "$DATABASE_URL" -c "INSERT INTO jobs(title, company, location, description) VALUES ('Admin Seed','Acme','Remote','Seeded by admin');"`
- List jobs: `psql "$DATABASE_URL" -c "SELECT id,title,company,location FROM jobs;"`

## Running locally
```bash
make format lint test
PORT=8080 go run ./cmd/server
```

## Docker and Compose
Build and run the API alongside Postgres and the mock AI (the app continues to operate in-memory, but the database is available for future persistence):
```bash
docker-compose up --build
```

The mock AI listens on port `18081` and returns deterministic JSON payloads. The API is exposed on `http://localhost:8080`.

## Sample prompts & fixtures
Base prompt used in code: `Extract JSON with title, company, location, description`

Example description and expected AI response shape:
```text
We are hiring a backend engineer at Golangjobsuz to build our API platform. Remote friendly.
```
AI output (JSON string):
```json
{"title":"Backend Engineer","company":"Golangjobsuz","location":"Remote","description":"Build the API platform"}
```

These fixtures can be fed to the mock AI or used in unit tests via `ai.MockClient`.
