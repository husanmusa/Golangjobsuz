package ai

import "context"

// Client represents an AI provider capable of generating responses.
type Client interface {
	GenerateResponse(ctx context.Context, prompt string) (string, error)
}

// NoopClient returns canned responses for development environments.
type NoopClient struct{}

// NewNoop returns a placeholder AI client.
func NewNoop() *NoopClient {
	return &NoopClient{}
}

// GenerateResponse returns a simple echo for development.
func (n *NoopClient) GenerateResponse(_ context.Context, prompt string) (string, error) {
	return "Echo: " + prompt, nil
}
