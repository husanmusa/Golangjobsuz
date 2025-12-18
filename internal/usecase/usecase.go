package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Golangjobsuz/bot/internal/ai"
	"github.com/Golangjobsuz/bot/internal/entities"
	"github.com/Golangjobsuz/bot/internal/repo"
)

// BotUseCase describes the operations the bot layer can trigger.
type BotUseCase interface {
	HandleMessage(ctx context.Context, msg *entities.Message) (string, error)
}

// Manager is a thin orchestrator for bot interactions.
type Manager struct {
	repositories repo.Storage
	aiClient     ai.Client
}

// NewManager wires dependencies into a bot use case manager.
func NewManager(repositories repo.Storage, aiClient ai.Client) *Manager {
	return &Manager{repositories: repositories, aiClient: aiClient}
}

// HandleMessage coordinates storing incoming messages, upserting users, and delegating to the AI client.
func (m *Manager) HandleMessage(ctx context.Context, msg *entities.Message) (string, error) {
	user := &entities.User{
		ID:        msg.ChatID,
		Username:  msg.Username,
		FirstName: msg.FirstName,
		LastName:  msg.LastName,
	}

	if m.repositories.Users != nil {
		if err := m.repositories.Users.Upsert(ctx, user); err != nil {
			return "", fmt.Errorf("upsert user: %w", err)
		}
	}

	if m.repositories.Messages != nil {
		copy := *msg
		if copy.Received.IsZero() {
			copy.Received = time.Now()
		}
		if err := m.repositories.Messages.Save(ctx, &copy); err != nil {
			return "", fmt.Errorf("save message: %w", err)
		}
	}

	prompt := strings.TrimSpace(msg.Text)
	if prompt == "" {
		prompt = "How can I help you today?"
	}

	if m.aiClient == nil {
		return prompt, nil
	}

	response, err := m.aiClient.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("generate response: %w", err)
	}

	return response, nil
}
