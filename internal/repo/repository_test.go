package repo

import (
	"context"
	"testing"
)

func TestInsertAndGet(t *testing.T) {
	repo := New()
	repo.InitSchema(context.Background())

	id, err := repo.Insert(context.Background(), &Job{Title: "t", Company: "c", Location: "l", Description: "d"})
	if err != nil {
		t.Fatalf("insert err: %v", err)
	}
	if id != 1 {
		t.Fatalf("expected id 1, got %d", id)
	}

	job, err := repo.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get err: %v", err)
	}
	if job.Title != "t" {
		t.Fatalf("unexpected title: %s", job.Title)
	}
}

func TestList(t *testing.T) {
	repo := New()
	repo.InitSchema(context.Background())
	ctx := context.Background()
	_, _ = repo.Insert(ctx, &Job{Title: "t", Company: "c", Location: "l", Description: "d"})

	jobs, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("list err: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected one job, got %d", len(jobs))
	}
}
