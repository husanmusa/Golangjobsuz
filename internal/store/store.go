package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// User represents a Telegram user interacting with the bot.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// RecruiterAccess tracks the approval/ban status for recruiters.
type RecruiterAccess struct {
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"` // pending, approved, banned
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
	Notes     string    `json:"notes"`
}

// Profile represents a candidate profile that can be searched.
type Profile struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Location     string    `json:"location"`
	Seniority    string    `json:"seniority"`
	Skills       []string  `json:"skills"`
	Summary      string    `json:"summary"`
	UpdatedAt    time.Time `json:"updated_at"`
	ContactEmail string    `json:"contact_email"`
	ContactPhone string    `json:"contact_phone"`
}

// Store holds all persistent data for the bot.
type Store struct {
	Users           map[string]User            `json:"users"`
	RecruiterAccess map[string]RecruiterAccess `json:"recruiter_access"`
	Profiles        map[string]Profile         `json:"profiles"`
	path            string                     `json:"-"`
}

// Load reads data from the given path, creating defaults if the file does not exist.
func Load(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("path is required")
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("mkdir data dir: %w", err)
		}
		s := defaultStore(path)
		if err := s.Save(); err != nil {
			return nil, err
		}
		return s, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read store: %w", err)
	}
	var store Store
	if err := json.Unmarshal(content, &store); err != nil {
		return nil, fmt.Errorf("unmarshal store: %w", err)
	}
	store.path = path
	store.ensureMaps()
	return &store, nil
}

// Save writes the current store state to disk.
func (s *Store) Save() error {
	if s.path == "" {
		return errors.New("store path missing")
	}
	s.ensureMaps()
	payload, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal store: %w", err)
	}
	if err := os.WriteFile(s.path, payload, 0o644); err != nil {
		return fmt.Errorf("write store: %w", err)
	}
	return nil
}

// ensureMaps initializes nil maps to avoid nil map panics.
func (s *Store) ensureMaps() {
	if s.Users == nil {
		s.Users = make(map[string]User)
	}
	if s.RecruiterAccess == nil {
		s.RecruiterAccess = make(map[string]RecruiterAccess)
	}
	if s.Profiles == nil {
		s.Profiles = make(map[string]Profile)
	}
}

// EnsureUserRole sets up a user with the provided role if they do not already exist.
func (s *Store) EnsureUserRole(userID, role string) {
	s.ensureMaps()
	u, ok := s.Users[userID]
	if !ok {
		u = User{ID: userID, CreatedAt: time.Now().UTC(), Name: userID}
	}
	u.Role = role
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	s.Users[userID] = u
}

func defaultStore(path string) *Store {
	now := time.Now().UTC()
	return &Store{
		Users: map[string]User{
			"u1":    {ID: "u1", Name: "Nargiza", Role: "candidate", CreatedAt: now},
			"u2":    {ID: "u2", Name: "Bekzod", Role: "recruiter", CreatedAt: now},
			"admin": {ID: "admin", Name: "Site Admin", Role: "admin", CreatedAt: now},
		},
		RecruiterAccess: map[string]RecruiterAccess{
			"u2": {UserID: "u2", Status: "approved", UpdatedAt: now, UpdatedBy: "admin", Notes: "default approved recruiter"},
		},
		Profiles: map[string]Profile{
			"p1": {
				ID:           "p1",
				Name:         "Akmal A.",
				Location:     "Tashkent",
				Seniority:    "mid",
				Skills:       []string{"golang", "postgres", "microservices"},
				Summary:      "Backend engineer with 4 years in fintech integrations.",
				UpdatedAt:    now.AddDate(0, 0, -2),
				ContactEmail: "akmal@example.com",
				ContactPhone: "+998901234567",
			},
			"p2": {
				ID:           "p2",
				Name:         "Sarvar S.",
				Location:     "Samarkand",
				Seniority:    "senior",
				Skills:       []string{"golang", "kubernetes", "grpc", "aws"},
				Summary:      "Senior Go dev leading cloud migrations.",
				UpdatedAt:    now.AddDate(0, 0, -8),
				ContactEmail: "sarvar@example.com",
				ContactPhone: "+998991112233",
			},
		},
		path: path,
	}
}
