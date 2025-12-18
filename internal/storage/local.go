package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage writes files to the local filesystem under a base directory.
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if basePath == "" {
		return nil, fmt.Errorf("base path must be provided")
	}
	return &LocalStorage{basePath: basePath}, nil
}

// Save writes the reader content to the given relative path and returns the absolute path.
func (s *LocalStorage) Save(ctx context.Context, relativePath string, r io.Reader) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	absPath := filepath.Join(s.basePath, relativePath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", fmt.Errorf("create directories: %w", err)
	}

	f, err := os.Create(absPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return absPath, nil
}
