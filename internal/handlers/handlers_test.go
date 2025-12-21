package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Golangjobsuz/golangjobsuz/internal/parser"
	"github.com/Golangjobsuz/golangjobsuz/internal/repo"
)

type stubPipeline struct{}

func (s *stubPipeline) Parse(_ context.Context, raw string) (*parser.Job, error) {
	return &parser.Job{Title: "t", Company: "c", Location: "l", Description: raw}, nil
}

func TestCreateJob(t *testing.T) {
	repository := repo.New()
	repository.InitSchema(context.Background())
	api := &API{Parser: &stubPipeline{}, Repo: repository}

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBuffer([]byte(`{"description":"d"}`)))
	rec := httptest.NewRecorder()
	api.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var job parser.Job
	_ = json.NewDecoder(rec.Body).Decode(&job)
	if job.ID != 1 || job.Description != "d" {
		t.Fatalf("unexpected job response: %+v", job)
	}
}

func TestListJobs(t *testing.T) {
	repository := repo.New()
	repository.InitSchema(context.Background())
	ctx := context.Background()
	_, _ = repository.Insert(ctx, &repo.Job{Title: "t", Company: "c", Location: "l", Description: "d"})
	api := &API{Parser: &stubPipeline{}, Repo: repository}

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()
	api.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
