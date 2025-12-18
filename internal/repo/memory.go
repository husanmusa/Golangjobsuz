package repo

import (
	"context"
	"sync"
	"time"

	"github.com/Golangjobsuz/bot/internal/entities"
)

// InMemoryUserRepository provides a thread-safe user store for prototyping.
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[int64]*entities.User
}

// NewInMemoryUserRepository constructs an empty in-memory user store.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{users: make(map[int64]*entities.User)}
}

// Upsert writes or updates a user record.
func (r *InMemoryUserRepository) Upsert(_ context.Context, user *entities.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	r.users[user.ID] = user
	return nil
}

// InMemoryMessageRepository provides a simple sink for messages.
type InMemoryMessageRepository struct {
	mu       sync.RWMutex
	messages []*entities.Message
}

// NewInMemoryMessageRepository constructs an empty message store.
func NewInMemoryMessageRepository() *InMemoryMessageRepository {
	return &InMemoryMessageRepository{messages: []*entities.Message{}}
}

// Save appends a message to the in-memory slice.
func (r *InMemoryMessageRepository) Save(_ context.Context, msg *entities.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, msg)
	return nil
}
