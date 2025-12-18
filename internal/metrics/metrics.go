package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Registry holds application metrics.
type Registry struct {
	UploadsTotal     atomic.Uint64
	UploadsBlocked   atomic.Uint64
	AISuccessTotal   atomic.Uint64
	AIFailedTotal    atomic.Uint64
	SubmissionsTotal atomic.Uint64
	TotalCostMicros  atomic.Uint64
	LatencyNanos     atomic.Uint64
	RequestsTotal    atomic.Uint64
}

// ObserveLatency adds the latency for a completed request.
func (r *Registry) ObserveLatency(d time.Duration) {
	r.LatencyNanos.Add(uint64(d.Nanoseconds()))
	r.RequestsTotal.Add(1)
}

// ObserveCost records the cost of an AI call in micro units to avoid floating point errors.
func (r *Registry) ObserveCost(micros uint64) {
	r.TotalCostMicros.Add(micros)
}

// ServeHTTP exposes metrics as JSON for easy scraping/logging.
func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	payload := map[string]any{
		"uploads_total":      r.UploadsTotal.Load(),
		"uploads_blocked":    r.UploadsBlocked.Load(),
		"ai_success_total":   r.AISuccessTotal.Load(),
		"ai_failed_total":    r.AIFailedTotal.Load(),
		"submissions_total":  r.SubmissionsTotal.Load(),
		"total_cost_micros":  r.TotalCostMicros.Load(),
		"requests_total":     r.RequestsTotal.Load(),
		"latency_average_ms": r.averageLatencyMillis(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func (r *Registry) averageLatencyMillis() float64 {
	reqs := r.RequestsTotal.Load()
	if reqs == 0 {
		return 0
	}
	nanos := r.LatencyNanos.Load()
	return float64(nanos) / float64(reqs) / 1e6
}
