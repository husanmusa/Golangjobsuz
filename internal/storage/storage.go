package storage

import (
	"context"
	"io"
)

// Backend allows storing binary or text content in persistent storage (local filesystem, S3, etc.).
type Backend interface {
	// Save writes content from r into the given relative path and returns the absolute or remote path where it was stored.
	Save(ctx context.Context, relativePath string, r io.Reader) (string, error)
}
