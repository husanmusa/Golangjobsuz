package search

import (
	"testing"
	"time"

	"github.com/Golangjobsuz/golangjobsuz/internal/store"
)

func TestSearchFiltersAndPagination(t *testing.T) {
	now := time.Now()
	profiles := map[string]store.Profile{
		"p1": {ID: "p1", Location: "Tashkent", Seniority: "mid", Skills: []string{"go", "grpc"}, UpdatedAt: now.Add(-2 * time.Hour)},
		"p2": {ID: "p2", Location: "Tashkent", Seniority: "senior", Skills: []string{"go", "aws"}, UpdatedAt: now.Add(-48 * time.Hour)},
		"p3": {ID: "p3", Location: "Samarkand", Seniority: "mid", Skills: []string{"python"}, UpdatedAt: now.Add(-24 * time.Hour)},
	}

	filters := Filters{Skills: []string{"go"}, Location: "Tashkent", MaxAgeDays: 3, PageSize: 1, Page: 2}
	results := SearchProfiles(profiles, filters)

	if results.Total != 2 {
		t.Fatalf("expected 2 results, got %d", results.Total)
	}
	if results.TotalPages != 2 || results.CurrentPage != 2 {
		t.Fatalf("expected two pages and current page 2, got pages=%d current=%d", results.TotalPages, results.CurrentPage)
	}
	if len(results.Results) != 1 || results.Results[0].Profile.ID != "p2" {
		t.Fatalf("expected p2 on second page, got %+v", results.Results)
	}
}

func TestRedaction(t *testing.T) {
	profile := store.Profile{ContactEmail: "hello@example.com", ContactPhone: "+998991234567"}
	res := RedactContact(profile)
	if res.RedactedEmail == "hello@example.com" || res.RedactedEmail == "" {
		t.Fatalf("email not redacted: %s", res.RedactedEmail)
	}
	if res.RedactedPhone == profile.ContactPhone || res.RedactedPhone == "" {
		t.Fatalf("phone not redacted: %s", res.RedactedPhone)
	}
}
