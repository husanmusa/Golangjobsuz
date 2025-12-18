package broadcast

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockSender struct {
	calls     int
	failUntil int
}

func (m *mockSender) Send(_ context.Context, channel, message string) error {
	m.calls++
	if channel == "" {
		return errors.New("missing channel")
	}
	if message == "" {
		return errors.New("empty message")
	}
	if m.calls <= m.failUntil {
		return errors.New("temporary failure")
	}
	return nil
}

type stubRepo struct {
	saved []BroadcastRecord
}

func (r *stubRepo) Save(_ context.Context, record BroadcastRecord) error {
	r.saved = append(r.saved, record)
	return nil
}

func (r *stubRepo) List(_ context.Context) ([]BroadcastRecord, error) {
	return r.saved, nil
}

func TestFormatCardIncludesSummaryAndFields(t *testing.T) {
	posting := JobPosting{
		Title:       "Backend Engineer",
		Company:     "ACME",
		Location:    "Remote",
		Salary:      "$4k",
		Experience:  "3y",
		Description: "Build services",
		Contact:     "@acmejobs",
	}

	summary := "Backend Engineer — Build services"
	card := FormatCard(posting, summary)
	for _, expected := range []string{
		"Backend Engineer",
		"ACME",
		summary,
		"Location: Remote",
		"Salary: $4k",
		"Experience: 3y",
		"Contact: @acmejobs",
	} {
		if !strings.Contains(card, expected) {
			t.Fatalf("expected %q in card: %s", expected, card)
		}
	}
}

func TestPostBroadcastDryRunRecordsWithoutSending(t *testing.T) {
	repo := &stubRepo{}
	sender := &mockSender{}
	svc := NewService(sender, repo, SimpleSummarizer{}, "#vacancies")

	posting := JobPosting{Title: "Backend", Company: "ACME"}
	record, err := svc.PostBroadcast(context.Background(), posting, Options{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if record.Status != StatusDryRun {
		t.Fatalf("expected dry run status, got %s", record.Status)
	}
	if sender.calls != 0 {
		t.Fatalf("expected sender not called, got %d", sender.calls)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected record saved")
	}
}

func TestPostBroadcastRetriesAndRecordsFailure(t *testing.T) {
	repo := &stubRepo{}
	sender := &mockSender{failUntil: 2}
	svc := NewService(sender, repo, SimpleSummarizer{}, "#vacancies")

	posting := JobPosting{Title: "Backend", Company: "ACME"}
	record, err := svc.PostBroadcast(context.Background(), posting, Options{MaxRetries: 2})
	if err == nil {
		t.Fatalf("expected error after retries")
	}

	if sender.calls != 2 {
		t.Fatalf("expected 2 attempts, got %d", sender.calls)
	}
	if record.Status != StatusFailed {
		t.Fatalf("expected failed status, got %s", record.Status)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected record saved")
	}
	if len(record.Errors) == 0 {
		t.Fatalf("expected errors recorded")
	}
}

func TestPostBroadcastSucceedsAfterRetry(t *testing.T) {
	repo := &stubRepo{}
	sender := &mockSender{failUntil: 1}
	svc := NewService(sender, repo, SimpleSummarizer{}, "#vacancies")

	posting := JobPosting{Title: "Backend", Company: "ACME"}
	record, err := svc.PostBroadcast(context.Background(), posting, Options{MaxRetries: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sender.calls != 2 {
		t.Fatalf("expected 2 attempts, got %d", sender.calls)
	}
	if record.Status != StatusSent {
		t.Fatalf("expected sent status, got %s", record.Status)
	}
	if record.LastSentAt == nil {
		t.Fatalf("expected LastSentAt set")
	}
}
