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
