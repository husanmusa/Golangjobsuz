package repo

import (
	"context"
	"errors"
	"sync"
)

// Job represents the persisted job schema matching the parser output.
type Job struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

// Repository stores job records in memory using a mutex for thread safety.
type Repository struct {
	mu    sync.Mutex
	next  int64
	jobs  map[int64]Job
	ready bool
}

// New constructs the repository.
func New() *Repository {
	return &Repository{jobs: make(map[int64]Job), next: 1}
}

// InitSchema mirrors the shape of database migrations and is retained for API compatibility.
func (r *Repository) InitSchema(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ready = true
	return nil
}

// Insert adds a new job.
func (r *Repository) Insert(_ context.Context, job *Job) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.ready {
		return 0, errors.New("schema not initialized")
	}
	if job == nil {
		return 0, errors.New("job is nil")
	}
	id := r.next
	r.next++
	copy := *job
	copy.ID = id
	r.jobs[id] = copy
	return id, nil
}

// Get fetches a job by ID.
func (r *Repository) Get(_ context.Context, id int64) (*Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[id]
	if !ok {
		return nil, errors.New("job not found")
	}
	return &job, nil
}

// List returns all jobs ordered by id.
func (r *Repository) List(_ context.Context) ([]Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	jobs := make([]Job, 0, len(r.jobs))
	for i := int64(1); i < r.next; i++ {
		if job, ok := r.jobs[i]; ok {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil

	"github.com/Golangjobsuz/bot/internal/entities"
)

// UserRepository persists Telegram user information.
type UserRepository interface {
	Upsert(ctx context.Context, user *entities.User) error
}

// MessageRepository persists incoming messages for auditing purposes.
type MessageRepository interface {
	Save(ctx context.Context, msg *entities.Message) error
}

// Storage aggregates all repositories the application depends on.
type Storage struct {
	Users    UserRepository
	Messages MessageRepository
}
