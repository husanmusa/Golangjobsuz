# Golangjobsuz Telegram Ingestion Bot

This bot accepts documents or direct links from Telegram, validates them, downloads the content, extracts text (including an optional OCR fallback for images), and stores both the raw file and extracted text in either local storage or S3.

## Features
- Telegram handler for document uploads and URL submissions with MIME type and size validation.
- Downloads files directly through the Telegram API with clear user-facing retry guidance.
- Text extraction for PDF and DOCX files, plus optional OCR via an HTTP endpoint for images.
- Pluggable storage backends (local filesystem or S3) for raw and extracted text outputs.
- Operation timeouts to prevent long-running tasks from blocking the bot.

## Running the bot
1. Ensure Go 1.20+ is installed.
2. Export the required environment variables:
   - `TELEGRAM_BOT_TOKEN` (required)
   - `LOCAL_STORAGE_PATH` (optional, defaults to `data/`)
   - `S3_BUCKET` and optional `S3_PREFIX` to use S3 instead of local storage.
3. Build and run:
   ```bash
   go run ./cmd/bot
   ```

## OCR configuration
The built-in OCR hook posts the image bytes to an HTTP endpoint. Configure `extract.Extractor{OCR: &extract.HttpOCRProvider{URL: "https://your-ocr-endpoint", Timeout: 30 * time.Second}}` when constructing the service to enable OCR for images.

## Testing and development notes
Network access to download Go modules may be restricted in some environments. If `go mod tidy` or `go test ./...` fails with proxy errors, ensure module downloads are allowed or use a module proxy that is accessible from your environment.
