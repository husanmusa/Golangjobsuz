package tests

import (
	"context"
	"testing"

	"github.com/example/golangjobsuz/internal/ai"
	"github.com/example/golangjobsuz/internal/parser"
	"github.com/example/golangjobsuz/internal/repo"
)

// Integration between parser and repository using the in-memory test database.
func TestPipelineToRepository(t *testing.T) {
	mock := &ai.MockClient{Responses: map[string]string{
		"prompt\n---\njob": `{"title":"Integration","company":"Example","location":"Remote","description":"Testing"}`,
	}}
	pipe := parser.NewPipeline(mock, "prompt")
	repository := repo.New()
	repository.InitSchema(context.Background())

	job, err := pipe.Parse(context.Background(), "job")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	id, err := repository.Insert(context.Background(), &repo.Job{
		Title:       job.Title,
		Company:     job.Company,
		Location:    job.Location,
		Description: job.Description,
	})
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	saved, err := repository.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if saved.Title != job.Title {
		t.Fatalf("unexpected job returned: %+v", saved)
	}
}
