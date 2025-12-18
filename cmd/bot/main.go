package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Golangjobsuz/bot/internal/ai"
	"github.com/Golangjobsuz/bot/internal/platform/config"
	"github.com/Golangjobsuz/bot/internal/platform/database"
	"github.com/Golangjobsuz/bot/internal/platform/httpclient"
	"github.com/Golangjobsuz/bot/internal/platform/logger"
	"github.com/Golangjobsuz/bot/internal/platform/metrics"
	"github.com/Golangjobsuz/bot/internal/repo"
	"github.com/Golangjobsuz/bot/internal/telegram"
	"github.com/Golangjobsuz/bot/internal/usecase"
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

	aiClient := ai.NewNoop()
	usecases := usecase.NewManager(repositories, aiClient)

	bot, err := telegram.New(cfg.Telegram.Token, usecases, logg)
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
