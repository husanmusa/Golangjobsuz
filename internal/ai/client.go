package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client describes the interface expected by the parsing pipeline.
type Client interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// HTTPClient is a minimal AI client that POSTs prompts to an HTTP endpoint.
type HTTPClient struct {
	Endpoint    string
	HTTP        *http.Client
	ContentType string
}

// NewHTTPClient builds an HTTP client with reasonable defaults.
func NewHTTPClient(endpoint string) *HTTPClient {
	return &HTTPClient{
		Endpoint:    endpoint,
		ContentType: "application/json",
		HTTP: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Complete sends a prompt to the configured AI endpoint and returns the response body text.
func (c *HTTPClient) Complete(ctx context.Context, prompt string) (string, error) {
	payload := map[string]string{"prompt": prompt}
	buf, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal prompt: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewBuffer(buf))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", c.ContentType)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("ai responded with status %d", resp.StatusCode)
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return body.Text, nil
}

// MockClient is a lightweight client for testing pipelines without hitting the network.
type MockClient struct {
	Responses map[string]string
}

// Complete returns canned responses configured on the mock.
func (m *MockClient) Complete(_ context.Context, prompt string) (string, error) {
	if m.Responses == nil {
		return "", fmt.Errorf("no mock responses configured")
	}
	resp, ok := m.Responses[prompt]
	if !ok {
		return "", fmt.Errorf("no mock response for prompt")
	}
	return resp, nil
}
