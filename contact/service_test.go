package contact

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockNotifier struct {
	seekerMessages []string
	adminMessages  []string
	failSeeker     bool
}

func (m *mockNotifier) NotifySeeker(_ context.Context, seekerContact string, message string) error {
	if m.failSeeker {
		return errors.New("seeker unavailable")
	}
	m.seekerMessages = append(m.seekerMessages, seekerContact+": "+message)
	return nil
}

func (m *mockNotifier) NotifyAdmin(_ context.Context, message string) error {
	m.adminMessages = append(m.adminMessages, message)
	return nil
}

type stubLogRepo struct {
	entries []LogEntry
}

func (r *stubLogRepo) Save(_ context.Context, entry LogEntry) error {
	r.entries = append(r.entries, entry)
	return nil
}

func (r *stubLogRepo) List(_ context.Context) ([]LogEntry, error) {
	return r.entries, nil
}

func TestHandleRequestSendsToSeeker(t *testing.T) {
	notifier := &mockNotifier{}
	repo := &stubLogRepo{}
	svc := NewService(notifier, repo)

	req := Request{
		RecruiterName:    "Rita",
		RecruiterCompany: "Talent Partners",
		RecruiterContact: "rita@example.com",
		Role:             "Backend Engineer",
		SeekerName:       "Sam",
		SeekerContact:    "@sam",
		Notes:            "Available this week",
	}

	entry, err := svc.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notifier.seekerMessages) != 1 {
		t.Fatalf("expected seeker notified")
	}
	if strings.Contains(notifier.seekerMessages[0], "rita@example.com") == false {
		t.Fatalf("expected recruiter contact in message")
	}
	if entry.ViaAdmin {
		t.Fatalf("expected direct delivery, got admin relay")
	}
	if !entry.Delivered {
		t.Fatalf("expected delivery flagged")
	}
}

func TestHandleRequestFallsBackToAdminRelay(t *testing.T) {
	notifier := &mockNotifier{failSeeker: true}
	repo := &stubLogRepo{}
	svc := NewService(notifier, repo)

	req := Request{RecruiterName: "Rita", RecruiterCompany: "Talent", Role: "Backend", SeekerContact: "@sam"}

	entry, err := svc.HandleRequest(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error returned to caller")
	}

	if len(notifier.seekerMessages) != 0 {
		t.Fatalf("expected seeker notification skipped")
	}
	if len(notifier.adminMessages) != 0 {
		t.Fatalf("admin relay should not be used when seeker contact present, even on failure")
	}
	if entry.Delivered {
		t.Fatalf("entry should mark failure")
	}
}

func TestHandleRequestUsesAdminWhenRequested(t *testing.T) {
	notifier := &mockNotifier{}
	repo := &stubLogRepo{}
	svc := NewService(notifier, repo)

	req := Request{RecruiterName: "Rita", RecruiterCompany: "Talent", Role: "Backend", UseAdminRelay: true}

	entry, err := svc.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notifier.adminMessages) != 1 {
		t.Fatalf("expected admin notified")
	}
	if !entry.ViaAdmin {
		t.Fatalf("entry should mark admin relay")
	}
}
