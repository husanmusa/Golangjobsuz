package extraction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient implements the AIClient interface for Google's Gemini API.
type GeminiClient struct {
	client    *genai.Client
	model     string
	maxTokens int
	retries   int
	logger    *log.Logger
	validator Validator
	cost      TokenCost
}

// NewGeminiClient builds a configured Gemini client with guardrails.
func NewGeminiClient(ctx context.Context, apiKey, model string, maxTokens, retries int, logger *log.Logger, validator Validator, cost TokenCost) (*GeminiClient, error) {
	cl, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}
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

	return &GeminiClient{
		client:    cl,
		model:     model,
		maxTokens: maxTokens,
		retries:   retries,
		logger:    logger,
		validator: validator,
		cost:      cost,
	}, nil
}

// Extract requests structured content from Gemini, validates it, and returns the draft payload.
func (c *GeminiClient) Extract(ctx context.Context, sourceText string) (Draft, error) {
	prompt := structuredPrompt(sourceText)
	var lastErr error
	backoff := 250 * time.Millisecond

	for attempt := 1; attempt <= c.retries; attempt++ {
		start := time.Now()
		model := c.client.GenerativeModel(c.model)
		model.ResponseMimeType = "application/json"
		model.GenerationConfig = &genai.GenerationConfig{MaxOutputTokens: int32(c.maxTokens)}
		model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text("You are a structured resume parser that outputs compact JSON.")}}

		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		latency := time.Since(start)

		if err != nil {
			lastErr = err
			c.logger.Printf("gemini attempt %d failed: %v", attempt, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			lastErr = errors.New("gemini returned no content")
			c.logger.Printf("gemini attempt %d empty content", attempt)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
		if !ok {
			lastErr = errors.New("gemini response not text")
			c.logger.Printf("gemini attempt %d unexpected part type", attempt)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		draft, err := c.toDraft(string(textPart), c.model)
		if err != nil {
			lastErr = err
			c.logger.Printf("gemini attempt %d validation failed: %v", attempt, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		c.logUsage(latency, resp.UsageMetadata)
		return draft, nil
	}

	return Draft{}, fmt.Errorf("gemini extraction failed after %d attempts: %w", c.retries, lastErr)
}

func (c *GeminiClient) toDraft(raw, model string) (Draft, error) {
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

func (c *GeminiClient) logUsage(latency time.Duration, usage *genai.UsageMetadata) {
	var promptTokens, outputTokens int32
	if usage != nil {
		promptTokens = usage.PromptTokenCount
		outputTokens = usage.CandidatesTokenCount
	}
	inputCost := (float64(promptTokens) / 1000.0) * c.cost.InputCost
	outputCost := (float64(outputTokens) / 1000.0) * c.cost.OutputCost
	c.logger.Printf("gemini model=%s latency_ms=%d prompt_tokens=%d completion_tokens=%d estimated_cost=%.6f", c.model, latency.Milliseconds(), promptTokens, outputTokens, inputCost+outputCost)
}
