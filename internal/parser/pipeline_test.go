package parser

import (
	"context"
	"testing"

	"github.com/Golangjobsuz/golangjobsuz/internal/ai"
)

func TestPipelineParse(t *testing.T) {
	mock := &ai.MockClient{Responses: map[string]string{"prompt\n---\nrole": `{"title":"Dev","company":"Acme","location":"Remote","description":"Build"}`}}
	pipe := NewPipeline(mock, "prompt")

	job, err := pipe.Parse(context.Background(), "role")
	if err != nil {
		t.Fatalf("expected parse to succeed: %v", err)
	}
	if job.Title != "Dev" || job.Company != "Acme" || job.Location != "Remote" {
		t.Fatalf("unexpected job: %+v", job)
	}
}

func TestPipelineRejectsEmpty(t *testing.T) {
	pipe := NewPipeline(&ai.MockClient{Responses: map[string]string{}}, "prompt")
	if _, err := pipe.Parse(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty input")
	}
}
