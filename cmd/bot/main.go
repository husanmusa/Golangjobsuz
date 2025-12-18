package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"Golangjobsuz/internal/extract"
	"Golangjobsuz/internal/ingest"
	"Golangjobsuz/internal/storage"
	"Golangjobsuz/internal/telegram"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("create bot: %v", err)
	}

	// Storage configuration
	var backend storage.Backend
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("load aws config: %v", err)
		}
		client := s3.NewFromConfig(cfg)
		backend, err = storage.NewS3Storage(client, bucket, os.Getenv("S3_PREFIX"))
		if err != nil {
			log.Fatalf("init s3 storage: %v", err)
		}
	} else {
		base := os.Getenv("LOCAL_STORAGE_PATH")
		if base == "" {
			base = "data"
		}
		backend, err = storage.NewLocalStorage(base)
		if err != nil {
			log.Fatalf("init local storage: %v", err)
		}
	}

	extractor := &extract.Extractor{}

	service := ingest.NewService(backend, extractor, ingest.Config{
		MaxFileSizeBytes: 20 * 1024 * 1024,
		AllowedMIMEs:     []string{"application/pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		StoreText:        true,
		OperationTimeout: 60 * time.Second,
	})

	handler := telegram.NewHandler(bot, service, 20*1024*1024, []string{"application/pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"})

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Printf("Bot started. Listening for messages...")
	for update := range updates {
		ctx := context.Background()
		if update.Message == nil {
			continue
		}
		response := handler.ProcessUpdate(ctx, update)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("send message: %v", err)
		}
	}
}
