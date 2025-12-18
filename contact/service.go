package contact

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Notifier delivers contact requests to seekers or admin relays.
type Notifier interface {
	NotifySeeker(ctx context.Context, seekerContact string, message string) error
	NotifyAdmin(ctx context.Context, message string) error
}

// Request captures the details of a recruiter contact request.
type Request struct {
	RecruiterName    string
	RecruiterCompany string
	RecruiterContact string
	Role             string
	SeekerName       string
	SeekerContact    string
	Notes            string
	UseAdminRelay    bool
}

// LogEntry stores a record of notifications we attempted.
type LogEntry struct {
	Request   Request
	Delivered bool
	ViaAdmin  bool
	Timestamp time.Time
	Error     string
}

// LogRepo persists contact request attempts.
type LogRepo interface {
	Save(ctx context.Context, entry LogEntry) error
	List(ctx context.Context) ([]LogEntry, error)
}

// Service coordinates routing recruiter requests to seekers or admin relays.
type Service struct {
	notifier Notifier
	repo     LogRepo
	clock    func() time.Time
}

// NewService constructs a contact service.
func NewService(notifier Notifier, repo LogRepo) *Service {
	return &Service{
		notifier: notifier,
		repo:     repo,
		clock:    time.Now,
	}
}

// HandleRequest routes the recruiter request to the seeker or admin relay.
func (s *Service) HandleRequest(ctx context.Context, req Request) (LogEntry, error) {
	message := formatMessage(req)
	entry := LogEntry{Request: req, Timestamp: s.clock()}

	var err error
	if req.UseAdminRelay || req.SeekerContact == "" {
		entry.ViaAdmin = true
		err = s.notifier.NotifyAdmin(ctx, message)
	} else {
		err = s.notifier.NotifySeeker(ctx, req.SeekerContact, message)
	}

	if err != nil {
		entry.Error = err.Error()
	} else {
		entry.Delivered = true
	}

	if saveErr := s.repo.Save(ctx, entry); saveErr != nil {
		return entry, fmt.Errorf("save contact log: %w", saveErr)
	}

	if err != nil {
		return entry, err
	}

	return entry, nil
}

func formatMessage(req Request) string {
	contact := req.RecruiterContact
	if contact == "" {
		contact = "(contact details not provided)"
	}

	return fmt.Sprintf(
		"Recruiter %s (%s) is interested in %s. Contact: %s. Notes: %s",
		req.RecruiterName,
		req.RecruiterCompany,
		req.Role,
		contact,
		req.Notes,
	)
}

// MemoryLogRepo provides an in-memory LogRepo implementation for tests and simple deployments.
type MemoryLogRepo struct {
	entries []LogEntry
	mu      sync.Mutex
}

// NewMemoryLogRepo constructs an empty in-memory log.
func NewMemoryLogRepo() *MemoryLogRepo {
	return &MemoryLogRepo{}
}

// Save records an entry in memory.
func (r *MemoryLogRepo) Save(_ context.Context, entry LogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
	return nil
}

// List returns all contact log entries.
func (r *MemoryLogRepo) List(_ context.Context) ([]LogEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]LogEntry, len(r.entries))
	copy(out, r.entries)
	return out, nil
}
