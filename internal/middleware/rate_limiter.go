package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}

// RateLimiter throttles requests per client key (usually IP) for uploads or AI calls.
type RateLimiter struct {
	limitPerSec float64
	burst       float64
	mu          sync.Mutex
	buckets     map[string]*bucket
}

func NewRateLimiter(requestsPerSec float64, burst int) *RateLimiter {
	return &RateLimiter{limitPerSec: requestsPerSec, burst: float64(burst), buckets: make(map[string]*bucket)}
}

// Middleware wraps handlers with rate limiting behavior.
func (r *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		key := clientKey(req)
		if !r.allow(key) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (r *RateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	bkt, ok := r.buckets[key]
	now := time.Now()
	if !ok {
		bkt = &bucket{tokens: r.burst - 1, last: now}
		r.buckets[key] = bkt
		return true
	}

	elapsed := now.Sub(bkt.last).Seconds()
	bkt.tokens += elapsed * r.limitPerSec
	if bkt.tokens > r.burst {
		bkt.tokens = r.burst
	}
	bkt.last = now

	if bkt.tokens < 1 {
		return false
	}

	bkt.tokens -= 1
	return true
}

func clientKey(req *http.Request) string {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

// Cleanup removes stale limiters to avoid unbounded growth.
func (r *RateLimiter) Cleanup(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	r.mu.Lock()
	for key, limiter := range r.buckets {
		if limiter.last.Before(cutoff) {
			delete(r.buckets, key)
		}
	}
	r.mu.Unlock()
}
