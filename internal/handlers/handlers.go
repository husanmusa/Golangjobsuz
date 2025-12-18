package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/example/golangjobsuz/internal/parser"
	"github.com/example/golangjobsuz/internal/repo"
)

// API bundles dependencies for HTTP handlers.
type API struct {
	Parser interface {
		Parse(context.Context, string) (*parser.Job, error)
	}
	Repo *repo.Repository
}

// Router returns an HTTP router with routes registered.
func (a *API) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			a.createJob(w, r)
		case http.MethodGet:
			a.listJobs(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func (a *API) createJob(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	job, err := a.Parser.Parse(r.Context(), payload.Description)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	id, err := a.Repo.Insert(r.Context(), &repo.Job{
		Title:       job.Title,
		Company:     job.Company,
		Location:    job.Location,
		Description: job.Description,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	job.ID = id
	writeJSON(w, http.StatusCreated, job)
}

func (a *API) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := a.Repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
