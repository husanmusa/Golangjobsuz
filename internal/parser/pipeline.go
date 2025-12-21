package parser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Golangjobsuz/golangjobsuz/internal/ai"
)

// Job represents a normalized job posting.
type Job struct {
	ID          int64  `json:"id,omitempty"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

// Pipeline wraps an AI client and provides deterministic parsing behavior.
type Pipeline struct {
	client ai.Client
	prompt string
}

// NewPipeline builds a parser configured with a reusable prompt.
func NewPipeline(client ai.Client, prompt string) *Pipeline {
	return &Pipeline{client: client, prompt: prompt}
}

// Parse collects a job description, sends a combined prompt to the AI service and decodes the response.
func (p *Pipeline) Parse(ctx context.Context, raw string) (*Job, error) {
	if raw == "" {
		return nil, fmt.Errorf("job description cannot be empty")
	}
	prompt := fmt.Sprintf("%s\n---\n%s", p.prompt, raw)
	text, err := p.client.Complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var job Job
	if err := json.Unmarshal([]byte(text), &job); err != nil {
		return nil, fmt.Errorf("parse job json: %w", err)
	}
	return &job, nil
}
