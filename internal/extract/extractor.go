package extract

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pdf "github.com/ledongthuc/pdf"
)

// Result captures both the extracted text and any warnings that occurred.
type Result struct {
	Text     string
	Warnings []string
}

// Extractor pulls text from a reader based on MIME type.
type Extractor struct {
	OCR OCRProvider
}

// OCRProvider defines the minimal behavior required to run OCR on images.
type OCRProvider interface {
	Recognize(ctx context.Context, image io.Reader) (string, error)
}

// ExtractText extracts text from the reader given its MIME type. It will attempt OCR for
// images if configured.
func (e *Extractor) ExtractText(ctx context.Context, mimeType string, r io.Reader) (Result, error) {
	switch {
	case strings.Contains(mimeType, "pdf"):
		return extractPDF(ctx, r)
	case strings.Contains(mimeType, "wordprocessingml") || strings.Contains(mimeType, "msword"):
		return extractDOCX(r)
	case strings.HasPrefix(mimeType, "image/"):
		if e.OCR == nil {
			return Result{Warnings: []string{"no OCR provider configured"}}, fmt.Errorf("cannot OCR image: provider not configured")
		}
		text, err := e.OCR.Recognize(ctx, r)
		return Result{Text: text}, err
	default:
		return Result{}, fmt.Errorf("unsupported mime type: %s", mimeType)
	}
}

func extractPDF(ctx context.Context, r io.Reader) (Result, error) {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		return Result{}, fmt.Errorf("buffer pdf: %w", err)
	}

	reader, err := pdf.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return Result{}, fmt.Errorf("open pdf: %w", err)
	}

	var builder strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		default:
		}

		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			return Result{}, fmt.Errorf("read pdf page %d: %w", i, err)
		}
		builder.WriteString(content)
		builder.WriteString("\n")
	}

	return Result{Text: builder.String()}, nil
}

func extractDOCX(r io.Reader) (Result, error) {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		return Result{}, fmt.Errorf("buffer docx: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return Result{}, fmt.Errorf("open docx zip: %w", err)
	}

	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}
	if docFile == nil {
		return Result{}, errors.New("document.xml not found in docx")
	}

	rc, err := docFile.Open()
	if err != nil {
		return Result{}, fmt.Errorf("open docx xml: %w", err)
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var builder strings.Builder
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Result{}, fmt.Errorf("parse docx: %w", err)
		}

		switch el := tok.(type) {
		case xml.CharData:
			text := strings.TrimSpace(string(el))
			if text != "" {
				builder.WriteString(text)
				builder.WriteString(" ")
			}
		}
	}

	return Result{Text: builder.String()}, nil
}

// HttpOCRProvider is a lightweight OCR implementation that calls an external HTTP endpoint.
// It is intentionally simple to keep optional OCR support dependency-free.
type HttpOCRProvider struct {
	Client  *http.Client
	URL     string
	Timeout time.Duration
}

func (p *HttpOCRProvider) Recognize(ctx context.Context, image io.Reader) (string, error) {
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: p.Timeout}
	}

	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.URL, image)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("ocr request failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read ocr response: %w", err)
	}
	return string(body), nil
}
