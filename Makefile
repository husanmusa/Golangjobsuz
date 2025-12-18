.PHONY: test lint format integration

format:
	gofmt -w ./

lint:
	go vet ./...

unit:
	go test ./internal/ai ./internal/parser ./internal/repo ./internal/handlers

integration:
	go test ./internal/tests -run TestPipelineToRepository -count=1

test:
	go test ./...
.PHONY: migrate-up migrate-down

# Run database migrations using the golang-migrate CLI.
# Set DATABASE_DSN in your environment or via an .env file before invoking.
migrate-up:
@[ -n "$(DATABASE_DSN)" ] || (echo "DATABASE_DSN is required" && exit 1)
migrate -path migrations -database "$(DATABASE_DSN)" up

migrate-down:
@[ -n "$(DATABASE_DSN)" ] || (echo "DATABASE_DSN is required" && exit 1)
migrate -path migrations -database "$(DATABASE_DSN)" down
