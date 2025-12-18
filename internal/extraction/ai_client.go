package extraction

import (
	"context"
)

// AIClient defines the behavior for any provider that can extract structured candidate
// profiles from unstructured text using an LLM.
type AIClient interface {
	Extract(ctx context.Context, sourceText string) (Draft, error)
}
