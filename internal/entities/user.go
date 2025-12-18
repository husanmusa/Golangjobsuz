package entities

import "time"

// User represents a Telegram user interacting with the bot.
type User struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
	CreatedAt time.Time
}
