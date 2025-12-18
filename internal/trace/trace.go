package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type contextKey string

const requestIDKey contextKey = "requestID"

// NewRequestID returns a random identifier for correlating traces/logs.
func NewRequestID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b[:])
}

// WithRequestID stores the given request ID in the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// FromContext retrieves a request ID if present.
func FromContext(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return "unknown"
}
