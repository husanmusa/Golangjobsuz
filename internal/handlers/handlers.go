package handlers

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"log/slog"

	"Golangjobsuz/internal/metrics"
	"Golangjobsuz/internal/notifier"
	"Golangjobsuz/internal/trace"
)

const (
	consentHeader     = "X-User-Consent"
	allowedCostMicros = 125000 // example cost ceiling per call in micros (0.125 currency units)
)

type App struct {
	Logger   *slog.Logger
	Metrics  *metrics.Registry
	Notifier *notifier.Notifier
}

var allowedMIMETypes = []string{
	"application/pdf",
	"text/plain",
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}

// UploadHandler enforces consent, scans file types, and records metrics.
func (a *App) UploadHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer a.trackLatency(start)

	requestID := trace.NewRequestID()
	ctx := trace.WithRequestID(r.Context(), requestID)
	r = r.WithContext(ctx)

	if consent := r.Header.Get(consentHeader); strings.ToLower(consent) != "true" {
		a.Logger.Info("consent required", "request_id", requestID)
		http.Error(w, "Consent required before processing submissions", http.StatusForbidden)
		return
	}
	a.Logger.Info("consent accepted", "request_id", requestID)

	file, header, err := r.FormFile("file")
	if err != nil {
		a.handleError(ctx, w, err, "missing file")
		return
	}
	defer file.Close()

	if disallowed := a.scanFileType(file, header.Filename); disallowed != "" {
		a.Metrics.UploadsBlocked.Add(1)
		a.Logger.Warn("blocked disallowed upload", "request_id", requestID, "reason", disallowed)
		http.Error(w, "Unsupported or dangerous file type", http.StatusBadRequest)
		return
	}

	a.Metrics.UploadsTotal.Add(1)
	a.Metrics.SubmissionsTotal.Add(1)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Upload accepted for processing"))
}

// AIHandler simulates an AI call with rate limiting and metrics.
func (a *App) AIHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer a.trackLatency(start)

	requestID := trace.NewRequestID()
	ctx := trace.WithRequestID(r.Context(), requestID)
	r = r.WithContext(ctx)

	prompt := strings.TrimSpace(r.FormValue("prompt"))
	if prompt == "" {
		a.handleError(ctx, w, errors.New("missing prompt"), "prompt missing")
		return
	}

	// Simulate call latency and cost measurement.
	time.Sleep(50 * time.Millisecond)
	a.Metrics.AISuccessTotal.Add(1)
	a.Metrics.ObserveCost(allowedCostMicros)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("AI call succeeded"))
}

// AuditHandler records admin or recruiter actions.
func (a *App) AuditHandler(w http.ResponseWriter, r *http.Request) {
	requestID := trace.NewRequestID()
	ctx := trace.WithRequestID(r.Context(), requestID)

	actor := r.FormValue("actor")
	role := r.FormValue("role")
	action := r.FormValue("action")
	if actor == "" || role == "" || action == "" {
		a.handleError(ctx, w, errors.New("missing audit fields"), "audit payload incomplete")
		return
	}

	a.Logger.Info("audit", "request_id", requestID, "actor", actor, "role", role, "action", action)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("audit recorded"))
}

func (a *App) trackLatency(start time.Time) {
	a.Metrics.ObserveLatency(time.Since(start))
}

func (a *App) scanFileType(file multipart.File, filename string) string {
	var sniff [512]byte
	n, _ := io.ReadFull(file, sniff[:])
	mimeType := http.DetectContentType(sniff[:n])
	_, _ = file.Seek(0, io.SeekStart)

	if !isAllowed(mimeType) {
		return mimeType
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".exe" || ext == ".bat" || ext == ".js" {
		return ext
	}
	return ""
}

func isAllowed(mimeType string) bool {
	for _, allowed := range allowedMIMETypes {
		if allowed == mimeType {
			return true
		}
	}
	if strings.HasPrefix(mimeType, "text/") {
		return true
	}
	return false
}

func (a *App) handleError(ctx context.Context, w http.ResponseWriter, err error, msg string) {
	requestID := trace.FromContext(ctx)
	a.Metrics.AIFailedTotal.Add(1)
	a.Logger.Error("request failed", "request_id", requestID, "error", err, "message", msg)
	a.Notifier.Alert(msg, "request_id", requestID)
	http.Error(w, msg, http.StatusBadRequest)
}
