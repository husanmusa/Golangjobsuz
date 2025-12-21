package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Golangjobsuz/golangjobsuz/broadcast"
	"github.com/Golangjobsuz/golangjobsuz/contact"
)

// consoleSender writes messages to stdout for demonstration purposes.
type consoleSender struct{}

func (consoleSender) Send(_ context.Context, channel, message string) error {
	fmt.Printf("[channel:%s]\n%s\n\n", channel, message)
	return nil
}

// consoleNotifier writes contact notifications to stdout.
type consoleNotifier struct{}

func (consoleNotifier) NotifySeeker(_ context.Context, seekerContact string, message string) error {
	fmt.Printf("[notify seeker %s]\n%s\n", seekerContact, message)
	return nil
}

func (consoleNotifier) NotifyAdmin(_ context.Context, message string) error {
	fmt.Printf("[notify admin]\n%s\n", message)
	return nil
}

func main() {
	ctx := context.Background()

	// Broadcast example
	repo := broadcast.NewFileRepo("data/broadcasts.json")
	svc := broadcast.NewService(consoleSender{}, repo, broadcast.SimpleSummarizer{}, "#vacancies")
	posting := broadcast.JobPosting{
		Title:       "Go Backend Engineer",
		Company:     "ExampleCo",
		Location:    "Remote",
		Salary:      "$5k-$6k",
		Experience:  "3+ years",
		Description: "Building APIs for a growing marketplace platform. Work with Go, Postgres, and Kubernetes.",
		Contact:     "talent@example.com",
	}

	if _, err := svc.PostBroadcast(ctx, posting, broadcast.Options{}); err != nil {
		log.Fatalf("broadcast failed: %v", err)
	}

	// Contact request example
	contactRepo := contact.NewMemoryLogRepo()
	contactSvc := contact.NewService(consoleNotifier{}, contactRepo)
	if _, err := contactSvc.HandleRequest(ctx, contact.Request{
		RecruiterName:    "Rita Recruiter",
		RecruiterCompany: "Talent Partners",
		RecruiterContact: "rita@example.com",
		Role:             posting.Title,
		SeekerName:       "Sam Seeker",
		SeekerContact:    "@samseeker",
		Notes:            "Available for a quick intro call",
	}); err != nil {
		log.Fatalf("contact flow failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
}
