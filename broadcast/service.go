package broadcast

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Sender posts formatted messages to a channel such as a vacancy feed.
type Sender interface {
	Send(ctx context.Context, channel, message string) error
}

// Summarizer generates a concise description of a job posting.
type Summarizer interface {
	Summarize(ctx context.Context, posting JobPosting) (string, error)
}

// JobPosting captures the details required for a broadcast card.
type JobPosting struct {
	Title       string
	Company     string
	Location    string
	Salary      string
	Experience  string
	Description string
	Contact     string
}

// RecordStatus represents the lifecycle state of a broadcast attempt.
type RecordStatus string

const (
	StatusPending RecordStatus = "pending"
	StatusSent    RecordStatus = "sent"
	StatusFailed  RecordStatus = "failed"
	StatusDryRun  RecordStatus = "dry-run"
)

// BroadcastRecord tracks the attempts to deliver a broadcast.
type BroadcastRecord struct {
	ID         string       `json:"id"`
	Job        JobPosting   `json:"job"`
	Summary    string       `json:"summary"`
	Channel    string       `json:"channel"`
	Status     RecordStatus `json:"status"`
	Attempts   int          `json:"attempts"`
	Errors     []string     `json:"errors"`
	DryRun     bool         `json:"dryRun"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	LastSentAt *time.Time   `json:"lastSentAt,omitempty"`
}

// Repo persists broadcast records for later inspection.
type Repo interface {
	Save(ctx context.Context, record BroadcastRecord) error
	List(ctx context.Context) ([]BroadcastRecord, error)
}

// Options controls how broadcasts are delivered.
type Options struct {
	DryRun     bool
	MaxRetries int
}

// Service coordinates formatting, sending, and tracking broadcasts.
type Service struct {
	sender     Sender
	repo       Repo
	summarizer Summarizer
	channel    string
	clock      func() time.Time
}

// NewService constructs a Service with sensible defaults.
func NewService(sender Sender, repo Repo, summarizer Summarizer, channel string) *Service {
	return &Service{
		sender:     sender,
		repo:       repo,
		summarizer: summarizer,
		channel:    channel,
		clock:      time.Now,
	}
}

// PostBroadcast formats a card and posts it to the configured channel.
func (s *Service) PostBroadcast(ctx context.Context, posting JobPosting, opts Options) (BroadcastRecord, error) {
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}

	summary, err := s.summarizer.Summarize(ctx, posting)
	if err != nil {
		return BroadcastRecord{}, fmt.Errorf("summarize: %w", err)
	}

	card := FormatCard(posting, summary)

	record := BroadcastRecord{
		ID:        fmt.Sprintf("%d", s.clock().UnixNano()),
		Job:       posting,
		Summary:   summary,
		Channel:   s.channel,
		Status:    StatusPending,
		CreatedAt: s.clock(),
		UpdatedAt: s.clock(),
		DryRun:    opts.DryRun,
	}

	if opts.DryRun {
		record.Status = StatusDryRun
		if err := s.repo.Save(ctx, record); err != nil {
			return record, fmt.Errorf("save dry-run record: %w", err)
		}
		return record, nil
	}

	for attempt := 1; attempt <= opts.MaxRetries; attempt++ {
		record.Attempts = attempt
		record.UpdatedAt = s.clock()

		sendErr := s.sender.Send(ctx, s.channel, card)
		if sendErr == nil {
			now := s.clock()
			record.LastSentAt = &now
			record.Status = StatusSent
			if err := s.repo.Save(ctx, record); err != nil {
				return record, fmt.Errorf("save sent record: %w", err)
			}
			return record, nil
		}

		record.Errors = append(record.Errors, sendErr.Error())
		if attempt == opts.MaxRetries {
			record.Status = StatusFailed
			if err := s.repo.Save(ctx, record); err != nil {
				return record, fmt.Errorf("save failed record: %w", err)
			}
			return record, sendErr
		}
	}

	return record, errors.New("unreachable")
}

// FormatCard builds the broadcast message using the AI summary plus key fields.
func FormatCard(posting JobPosting, summary string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%s* at *%s*\n", posting.Title, posting.Company)
	if summary != "" {
		fmt.Fprintf(&b, "%s\n\n", summary)
	}

	if posting.Location != "" {
		fmt.Fprintf(&b, "• Location: %s\n", posting.Location)
	}
	if posting.Salary != "" {
		fmt.Fprintf(&b, "• Salary: %s\n", posting.Salary)
	}
	if posting.Experience != "" {
		fmt.Fprintf(&b, "• Experience: %s\n", posting.Experience)
	}
	if posting.Description != "" {
		fmt.Fprintf(&b, "• Details: %s\n", posting.Description)
	}
	if posting.Contact != "" {
		fmt.Fprintf(&b, "• Contact: %s\n", posting.Contact)
	}

	return strings.TrimSpace(b.String())
}

// SimpleSummarizer provides a lightweight, deterministic summary for environments
// without an LLM integration.
type SimpleSummarizer struct{}

// Summarize condenses the description, preferring the first sentence or a trim.
func (SimpleSummarizer) Summarize(_ context.Context, posting JobPosting) (string, error) {
	desc := strings.TrimSpace(posting.Description)
	if desc == "" {
		return "", nil
	}

	parts := strings.Split(desc, ". ")
	first := strings.TrimSpace(parts[0])
	if len(first) > 180 {
		first = first[:180] + "…"
	}

	return fmt.Sprintf("%s — %s", posting.Title, first), nil
}

// FileRepo stores broadcasts in a JSON file for auditing.
type FileRepo struct {
	path string
	mu   sync.Mutex
}

// NewFileRepo constructs a repository rooted at the supplied path.
func NewFileRepo(path string) *FileRepo {
	return &FileRepo{path: path}
}

// Save appends or updates a record in the JSON store.
func (r *FileRepo) Save(_ context.Context, record BroadcastRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.readAll()
	if err != nil {
		return err
	}

	updated := false
	for i, rec := range records {
		if rec.ID == record.ID {
			records[i] = record
			updated = true
			break
		}
	}
	if !updated {
		records = append(records, record)
	}

	return r.writeAll(records)
}

// List retrieves all broadcast records.
func (r *FileRepo) List(_ context.Context) ([]BroadcastRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.readAll()
}

func (r *FileRepo) readAll() ([]BroadcastRecord, error) {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []BroadcastRecord{}, nil
		}
		return nil, err
	}

	var records []BroadcastRecord
	if len(data) == 0 {
		return []BroadcastRecord{}, nil
	}

	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *FileRepo) writeAll(records []BroadcastRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.path, data, 0o644)
}
