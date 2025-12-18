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
