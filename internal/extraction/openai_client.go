package extraction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the AIClient interface for the OpenAI API.
type OpenAIClient struct {
	client     *openai.Client
	model      string
	maxTokens  int
	retries    int
	logger     *log.Logger
	validator  Validator
	costConfig map[string]TokenCost
}

// TokenCost represents input/output token prices per 1K tokens.
type TokenCost struct {
	InputCost  float64
	OutputCost float64
}

// Validator abstracts validation to ease testing.
type Validator interface {
	Struct(s interface{}) error
}

// NewOpenAIClient builds a configured OpenAI client with guardrails.
func NewOpenAIClient(apiKey, model string, maxTokens, retries int, logger *log.Logger, validator Validator, costConfig map[string]TokenCost) *OpenAIClient {
	if logger == nil {
		logger = log.Default()
	}
	if validator == nil {
		validator = defaultValidator()
	}
	if retries < 1 {
		retries = 3
	}
	if maxTokens == 0 {
		maxTokens = 512
	}
	return &OpenAIClient{
		client:     openai.NewClient(apiKey),
		model:      model,
		maxTokens:  maxTokens,
		retries:    retries,
		logger:     logger,
		validator:  validator,
		costConfig: costConfig,
	}
}

// Extract requests a structured completion, validates it against the schema, and
// returns a Draft enriched with metadata.
func (c *OpenAIClient) Extract(ctx context.Context, sourceText string) (Draft, error) {
	prompt := structuredPrompt(sourceText)
	var lastErr error
	backoff := 250 * time.Millisecond

	for attempt := 1; attempt <= c.retries; attempt++ {
		start := time.Now()
		resp, err := c.client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model: c.model,
				Messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleSystem, Content: "You are a structured resume parser that outputs compact JSON."},
					{Role: openai.ChatMessageRoleUser, Content: prompt},
				},
				ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
				MaxTokens:      c.maxTokens,
			},
		)
		latency := time.Since(start)

		if err != nil {
			lastErr = err
			c.logger.Printf("openai attempt %d failed: %v", attempt, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if len(resp.Choices) == 0 {
			lastErr = errors.New("openai returned no choices")
			c.logger.Printf("openai attempt %d empty choices", attempt)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		content := resp.Choices[0].Message.Content
		draft, err := c.toDraft(content, resp.Model)
		if err != nil {
			lastErr = err
			c.logger.Printf("openai attempt %d validation failed: %v", attempt, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		c.logUsage(latency, resp.Usage)
		return draft, nil
	}

	return Draft{}, fmt.Errorf("openai extraction failed after %d attempts: %w", c.retries, lastErr)
}

func (c *OpenAIClient) toDraft(raw, model string) (Draft, error) {
	var profile CandidateProfile
	if err := json.Unmarshal([]byte(raw), &profile); err != nil {
		return Draft{}, fmt.Errorf("decode: %w", err)
	}

	if err := c.validator.Struct(profile); err != nil {
		return Draft{}, fmt.Errorf("validation: %w", err)
	}

	return Draft{
		Profile:     profile,
		RawResponse: raw,
		Model:       model,
		ExtractedAt: time.Now(),
	}, nil
}

func (c *OpenAIClient) logUsage(latency time.Duration, usage openai.Usage) {
	cost, ok := c.costConfig[c.model]
	if !ok {
		cost = TokenCost{}
	}
	inputCost := (float64(usage.PromptTokens) / 1000.0) * cost.InputCost
	outputCost := (float64(usage.CompletionTokens) / 1000.0) * cost.OutputCost
	c.logger.Printf("openai model=%s latency_ms=%d prompt_tokens=%d completion_tokens=%d estimated_cost=%.6f", c.model, latency.Milliseconds(), usage.PromptTokens, usage.CompletionTokens, inputCost+outputCost)
}
