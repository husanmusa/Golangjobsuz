# Golangjobsuz

A minimal API that parses job descriptions via an AI client and stores normalized jobs (in-memory by default) with a Postgres-compatible schema.

## Quick start
```bash
make format lint test
PORT=8080 DATABASE_URL=postgres://postgres:postgres@localhost:5432/golangjobs?sslmode=disable go run ./cmd/server
```

Run integration tests (requires Docker):
```bash
make integration
```

See [docs/RUNBOOK.md](docs/RUNBOOK.md) for environment variables, admin commands, and sample AI fixtures.
