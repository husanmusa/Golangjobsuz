package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry bundles Prometheus metric registration and HTTP exposure.
type Registry struct {
	registerer *prometheus.Registry
	handler    http.Handler
}

// New creates a new isolated Prometheus registry ready for instrumentation.
func New() *Registry {
	reg := prometheus.NewRegistry()
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return &Registry{registerer: reg, handler: handler}
}

// Registerer exposes the Prometheus registerer for adding metrics.
func (r *Registry) Registerer() prometheus.Registerer {
	return r.registerer
}

// Handler returns the HTTP handler exposing the metrics endpoint.
func (r *Registry) Handler() http.Handler {
	return r.handler
}
