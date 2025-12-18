package entities

import "time"

// Message represents an incoming Telegram message of interest to the bot.
type Message struct {
	ChatID    int64
	Text      string
	Username  string
	FirstName string
	LastName  string
	Received  time.Time
}
