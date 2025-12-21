package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Golangjobsuz/golangjobsuz/internal/ingest"
)

// Handler processes Telegram updates for document and link ingestion.
type Handler struct {
	bot        *tgbotapi.BotAPI
	service    *ingest.Service
	maxSize    int64
	allowed    []string
	httpClient *http.Client
}

func NewHandler(bot *tgbotapi.BotAPI, service *ingest.Service, maxSize int64, allowed []string) *Handler {
	return &Handler{bot: bot, service: service, maxSize: maxSize, allowed: allowed, httpClient: &http.Client{Timeout: 20 * time.Second}}
}

// ProcessUpdate validates the incoming update, downloads the payload, and ingests it. The returned
// string is safe to show to the user.
func (h *Handler) ProcessUpdate(ctx context.Context, update tgbotapi.Update) string {
	if update.Message == nil {
		return "Send a document or a link to process."
	}

	if doc := update.Message.Document; doc != nil {
		return h.handleDocument(ctx, update.Message.Chat.ID, doc)
	}

	if url := extractLink(update.Message); url != "" {
		return h.handleURL(ctx, update.Message.Chat.ID, url)
	}

	return "Unsupported message type. Please send a document or a direct link."
}

func (h *Handler) handleDocument(ctx context.Context, chatID int64, doc *tgbotapi.Document) string {
	if h.maxSize > 0 && doc.FileSize > h.maxSize {
		return fmt.Sprintf("File too large (max %d bytes).", h.maxSize)
	}
	if len(h.allowed) > 0 && !contains(h.allowed, doc.MimeType) {
		return fmt.Sprintf("Unsupported mime type: %s", doc.MimeType)
	}

	tgFile, err := h.bot.GetFile(tgbotapi.FileConfig{FileID: doc.FileID})
	if err != nil {
		return "Could not retrieve file metadata. Please try again."
	}

	dlURL := tgFile.Link(h.bot.Token)
	resp, err := h.httpClient.Get(dlURL)
	if err != nil {
		return "Download failed, please retry."
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "Download failed due to Telegram response. Try again later."
	}

	name := doc.FileName
	if name == "" {
		name = doc.FileID
	}

	output, err := h.service.Ingest(ctx, ingest.InputFile{
		Name:     name,
		MIMEType: doc.MimeType,
		Size:     int64(doc.FileSize),
		Content:  resp.Body,
	})
	if err != nil {
		return userFacingError(err)
	}

	if output.TextLocation != "" {
		return fmt.Sprintf("Stored file at %s and text at %s", output.RawLocation, output.TextLocation)
	}
	return fmt.Sprintf("Stored file at %s. Extracted %d bytes of text.", output.RawLocation, len(output.Extracted.Text))
}

func (h *Handler) handleURL(ctx context.Context, chatID int64, url string) string {
	resp, err := h.httpClient.Get(url)
	if err != nil {
		return "Could not download the provided link."
	}
	defer resp.Body.Close()

	head, err := io.ReadAll(io.LimitReader(resp.Body, 512))
	if err != nil {
		return "Failed to read remote content."
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = http.DetectContentType(head)
	}

	name := deriveNameFromURL(url)
	output, err := h.service.Ingest(ctx, ingest.InputFile{
		Name:     name,
		MIMEType: mime,
		Size:     resp.ContentLength,
		Content:  io.MultiReader(bytes.NewReader(head), resp.Body),
	})
	if err != nil {
		return userFacingError(err)
	}

	return fmt.Sprintf("Stored link content at %s. Extracted %d bytes of text.", output.RawLocation, len(output.Extracted.Text))
}

func extractLink(msg *tgbotapi.Message) string {
	if msg.Text == "" {
		return ""
	}
	if strings.HasPrefix(msg.Text, "http://") || strings.HasPrefix(msg.Text, "https://") {
		return msg.Text
	}
	for _, entity := range msg.Entities {
		if entity.Type == "url" || entity.Type == "text_link" {
			return msg.Text[entity.Offset : entity.Offset+entity.Length]
		}
	}
	return ""
}

func deriveNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "downloaded_file"
	}
	last := parts[len(parts)-1]
	if last == "" {
		return "downloaded_file"
	}
	return last
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if strings.EqualFold(v, value) {
			return true
		}
	}
	return false
}

func userFacingError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "Processing timed out. Please try again with a smaller file."
	}
	return fmt.Sprintf("Processing failed: %v. Please retry.", err)
}
