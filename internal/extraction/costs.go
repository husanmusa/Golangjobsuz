package extraction

// DefaultOpenAICosts lists input/output pricing per 1K tokens for common models.
var DefaultOpenAICosts = map[string]TokenCost{
	"gpt-4o-mini":      {InputCost: 0.00015, OutputCost: 0.0006},
	"gpt-4o-mini-2024": {InputCost: 0.00015, OutputCost: 0.0006},
	"gpt-4o":           {InputCost: 0.005, OutputCost: 0.015},
	"gpt-3.5-turbo":    {InputCost: 0.0005, OutputCost: 0.0015},
}

// DefaultGeminiCost provides an optional fallback cost configuration.
var DefaultGeminiCost = TokenCost{InputCost: 0.00035, OutputCost: 0.00105}
