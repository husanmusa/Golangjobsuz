package ingest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/Golangjobsuz/golangjobsuz/internal/extract"
	"github.com/Golangjobsuz/golangjobsuz/internal/storage"
)

// Config controls validation, timeouts, and allowed file types.
type Config struct {
	MaxFileSizeBytes int64
	AllowedMIMEs     []string
	StoreText        bool
	OperationTimeout time.Duration
}

// Service coordinates validation, storage, and text extraction.
type Service struct {
	store     storage.Backend
	extractor *extract.Extractor
	config    Config
}

func NewService(store storage.Backend, extractor *extract.Extractor, config Config) *Service {
	return &Service{store: store, extractor: extractor, config: config}
}

// InputFile describes a file to ingest.
type InputFile struct {
	Name     string
	MIMEType string
	Size     int64
	Content  io.Reader
}

// Output holds references to stored data.
type Output struct {
	RawLocation  string
	TextLocation string
	Extracted    extract.Result
}

// Ingest validates, stores, and extracts text from the incoming file.
func (s *Service) Ingest(ctx context.Context, file InputFile) (Output, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.OperationTimeout)
	defer cancel()

	if err := s.validate(file); err != nil {
		return Output{}, err
	}

	buf := &bytes.Buffer{}
	tee := io.TeeReader(file.Content, buf)

	rawPath := filepath.Join(time.Now().Format("2006/01/02"), file.Name)
	rawLocation, err := s.store.Save(ctx, rawPath, tee)
	if err != nil {
		return Output{}, fmt.Errorf("store raw file: %w", err)
	}

	result, err := s.extractor.ExtractText(ctx, file.MIMEType, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return Output{RawLocation: rawLocation, Extracted: result}, fmt.Errorf("extract text: %w", err)
	}

	var textLocation string
	if s.config.StoreText {
		textPath := strings.TrimSuffix(rawPath, filepath.Ext(rawPath)) + ".txt"
		textLocation, err = s.store.Save(ctx, textPath, strings.NewReader(result.Text))
		if err != nil {
			return Output{RawLocation: rawLocation, Extracted: result}, fmt.Errorf("store text: %w", err)
		}
	}

	return Output{RawLocation: rawLocation, TextLocation: textLocation, Extracted: result}, nil
}

func (s *Service) validate(file InputFile) error {
	if s.config.MaxFileSizeBytes > 0 && file.Size > s.config.MaxFileSizeBytes {
		return fmt.Errorf("file size %d exceeds max %d", file.Size, s.config.MaxFileSizeBytes)
	}

	if len(s.config.AllowedMIMEs) == 0 {
		return nil
	}

	for _, allowed := range s.config.AllowedMIMEs {
		if strings.EqualFold(allowed, file.MIMEType) {
			return nil
		}
	}
	return fmt.Errorf("mime type %s is not allowed", file.MIMEType)
}
