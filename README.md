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
