package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Golangjobsuz/golangjobsuz/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil {
		log.Println("server stopped:", err)
		os.Exit(1)
	}
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/Golangjobsuz/golangjobsuz/internal/handlers"
	"github.com/Golangjobsuz/golangjobsuz/internal/logging"
	"github.com/Golangjobsuz/golangjobsuz/internal/metrics"
	"github.com/Golangjobsuz/golangjobsuz/internal/middleware"
	"github.com/Golangjobsuz/golangjobsuz/internal/notifier"
)

func main() {
	logger := logging.NewLogger()
	metricsRegistry := &metrics.Registry{}
	notif := notifier.New(logger)

	app := &handlers.App{Logger: logger, Metrics: metricsRegistry, Notifier: notif}

	rateLimiter := middleware.NewRateLimiter(5, 5)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	mux.Handle("/metrics", metricsRegistry)
	mux.Handle("/upload", rateLimiter.Middleware(http.HandlerFunc(app.UploadHandler)))
	mux.Handle("/ai", rateLimiter.Middleware(http.HandlerFunc(app.AIHandler)))
	mux.Handle("/admin/audit", http.HandlerFunc(app.AuditHandler))

	server := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(logger, mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info("server starting", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request complete", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
