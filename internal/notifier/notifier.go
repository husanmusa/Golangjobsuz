package notifier

import "log/slog"

// Notifier emits alerts to operators. Here we log to stderr, but this can be
// wired to email/Slack/etc.
type Notifier struct {
	log *slog.Logger
}

func New(log *slog.Logger) *Notifier {
	return &Notifier{log: log}
}

// Alert logs an operational alert.
func (n *Notifier) Alert(msg string, attrs ...any) {
	n.log.Warn("alert", append([]any{"message", msg}, attrs...)...)
}
