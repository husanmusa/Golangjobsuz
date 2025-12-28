package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Golangjobsuz/golangjobsuz/internal/ai"
	"github.com/Golangjobsuz/golangjobsuz/internal/extract"
	"github.com/Golangjobsuz/golangjobsuz/internal/ingest"
	"github.com/Golangjobsuz/golangjobsuz/internal/platform/config"
	"github.com/Golangjobsuz/golangjobsuz/internal/platform/database"
	"github.com/Golangjobsuz/golangjobsuz/internal/platform/httpclient"
	"github.com/Golangjobsuz/golangjobsuz/internal/platform/logger"
	"github.com/Golangjobsuz/golangjobsuz/internal/platform/metrics"
	"github.com/Golangjobsuz/golangjobsuz/internal/repo"
	"github.com/Golangjobsuz/golangjobsuz/internal/storage"
	"github.com/Golangjobsuz/golangjobsuz/internal/telegram"
	"github.com/Golangjobsuz/golangjobsuz/internal/usecase"
)

const (
	defaultStoragePath   = "data/ingest"
	defaultMaxFileBytes  = 10 * 1024 * 1024
	defaultTimeoutSecond = 30
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logg := logger.New(cfg.AppName, cfg.Environment)

	dbPool, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		logg.Fatal().Err(err).Msg("connect database")
	}
	if dbPool != nil {
		defer dbPool.Close()
		logg.Info().Msg("database connection pool initialized")
	}

	metricsRegistry := metrics.New()
	go func() {
		srv := &http.Server{
			Addr:    cfg.Metrics.Address,
			Handler: metricsRegistry.Handler(),
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Error().Err(err).Msg("metrics server failed")
		}
	}()

	httpClient := httpclient.New(cfg.HTTP.TimeoutSeconds, cfg.HTTP.MaxRetries)
	_ = httpClient // currently unused but prepared for integrations

	repositories := repo.Storage{
		Users:    repo.NewInMemoryUserRepository(),
		Messages: repo.NewInMemoryMessageRepository(),
	}

	ingestService, maxSize, allowed, err := buildIngestService()
	if err != nil {
		logg.Fatal().Err(err).Msg("initialize ingest service")
	}

	aiClient := ai.NewNoop()
	usecases := usecase.NewManager(repositories, aiClient)

	bot, err := telegram.New(cfg.Telegram.Token, usecases, logg, ingestService, maxSize, allowed)
	if err != nil {
		logg.Fatal().Err(err).Msg("initialize telegram bot")
	}

	if err := bot.Start(ctx); err != nil {
		if err == context.Canceled {
			logg.Info().Msg("bot context canceled; exiting")
			os.Exit(0)
		}
		logg.Error().Err(err).Msg("bot stopped with error")
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}
}

func buildIngestService() (*ingest.Service, int64, []string, error) {
	localPath := getenv("TEMP_STORAGE_PATH", defaultStoragePath)
	store, err := storage.NewLocalStorage(localPath)
	if err != nil {
		return nil, 0, nil, err
	}

	extractor := &extract.Extractor{}
	maxBytes := getenvInt64("MAX_FILE_BYTES", defaultMaxFileBytes)
	timeout := time.Duration(getenvInt("REQUEST_TIMEOUT_SECONDS", defaultTimeoutSecond)) * time.Second

	allowed := defaultAllowedMIMEs()
	service := ingest.NewService(store, extractor, ingest.Config{
		MaxFileSizeBytes: maxBytes,
		AllowedMIMEs:     allowed,
		StoreText:        true,
		OperationTimeout: timeout,
	})
	return service, maxBytes, allowed, nil
}

func defaultAllowedMIMEs() []string {
	if value := strings.TrimSpace(os.Getenv("MAX_FILE_TYPES")); value != "" {
		return normalizeFileTypes(strings.Split(value, ","))
	}
	return []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
}

func normalizeFileTypes(items []string) []string {
	var allowed []string
	for _, item := range items {
		entry := strings.TrimSpace(item)
		if entry == "" {
			continue
		}
		if strings.HasPrefix(entry, ".") {
			switch strings.ToLower(entry) {
			case ".pdf":
				allowed = append(allowed, "application/pdf")
			case ".doc":
				allowed = append(allowed, "application/msword")
			case ".docx":
				allowed = append(allowed, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
			default:
				continue
			}
			continue
		}
		allowed = append(allowed, entry)
	}
	return allowed
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
