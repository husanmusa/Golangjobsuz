package repo

import (
	"context"

	"github.com/Golangjobsuz/bot/internal/entities"
)

// UserRepository persists Telegram user information.
type UserRepository interface {
	Upsert(ctx context.Context, user *entities.User) error
}

// MessageRepository persists incoming messages for auditing purposes.
type MessageRepository interface {
	Save(ctx context.Context, msg *entities.Message) error
}

// Storage aggregates all repositories the application depends on.
type Storage struct {
	Users    UserRepository
	Messages MessageRepository
}
