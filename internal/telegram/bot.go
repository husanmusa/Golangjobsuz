package telegram

import (
	"context"
	"errors"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Golangjobsuz/golangjobsuz/internal/entities"
	"github.com/Golangjobsuz/golangjobsuz/internal/platform/logger"
	"github.com/Golangjobsuz/golangjobsuz/internal/usecase"
)

// Bot encapsulates Telegram-specific wiring and lifecycle management.
type Bot struct {
	api      *tgbotapi.BotAPI
	usecases usecase.BotUseCase
	logger   logger.Logger
	client   *http.Client
}

// New constructs a Bot with the provided token and dependencies.
func New(token string, usecases usecase.BotUseCase, log logger.Logger) (*Bot, error) {
	if token == "" {
		return nil, errors.New("telegram token is required")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	api.Client.Timeout = 60 * time.Second

	return &Bot{api: api, usecases: usecases, logger: log, client: api.Client}, nil
}

// Start begins polling for updates and processing incoming messages.
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info().Msg("telegram bot starting")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := b.api.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info().Msg("telegram bot shutting down")
			return ctx.Err()
		case update, ok := <-updates:
			if !ok {
				return errors.New("update channel closed")
			}
			b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	username := ""
	firstName := ""
	lastName := ""
	if update.Message.From != nil {
		username = update.Message.From.UserName
		firstName = update.Message.From.FirstName
		lastName = update.Message.From.LastName
	}

	msg := &entities.Message{
		ChatID:    update.Message.Chat.ID,
		Text:      update.Message.Text,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Received:  time.Unix(int64(update.Message.Date), 0),
	}

	response, err := b.usecases.HandleMessage(ctx, msg)
	if err != nil {
		b.logger.Error().Err(err).Msg("handle message")
		response = "Something went wrong. Please try again."
	}

	reply := tgbotapi.NewMessage(msg.ChatID, response)
	if _, err := b.api.Send(reply); err != nil {
		b.logger.Error().Err(err).Msg("send telegram message")
	}
}
