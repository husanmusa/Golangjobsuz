package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Golangjobsuz/golangjobsuz/internal/ai"
	"github.com/Golangjobsuz/golangjobsuz/internal/config"
	"github.com/Golangjobsuz/golangjobsuz/internal/handlers"
	"github.com/Golangjobsuz/golangjobsuz/internal/parser"
	"github.com/Golangjobsuz/golangjobsuz/internal/repo"
)

// Run bootstraps dependencies and starts the HTTP server.
func Run(ctx context.Context) error {
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}

	repository := repo.New()
	if err := repository.InitSchema(ctx); err != nil {
		return err
	}

	aiClient := ai.NewHTTPClient(cfg.AIEndpoint)
	pipeline := parser.NewPipeline(aiClient, "Extract JSON with title, company, location, description")
	api := &handlers.API{Parser: pipeline, Repo: repository}

	router := api.Router()
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("server listening on %s", srv.Addr)
	return srv.ListenAndServe()
}
