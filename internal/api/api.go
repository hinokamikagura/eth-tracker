// Package api provides HTTP API endpoints for the eth-tracker service.
// Currently exposes Prometheus metrics and service statistics.
package api

import (
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/uptrace/bunrouter"
)

const (
	// metricsPath is the HTTP path for Prometheus metrics endpoint
	metricsPath = "/metrics"
	// statsPath is the HTTP path for service statistics endpoint
	statsPath = "/stats"
	// healthPath is the HTTP path for health check endpoint
	healthPath = "/health"
)

// New creates a new HTTP router with all API endpoints registered.
func New() *bunrouter.Router {
	router := bunrouter.New()

	router.GET(metricsPath, metricsHandler())
	router.GET(statsPath, statsHandler())
	router.GET(healthPath, healthHandler())

	return router
}

// metricsHandler returns a handler that serves Prometheus metrics.
func metricsHandler() bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, _ bunrouter.Request) error {
		metrics.WritePrometheus(w, true)
		return nil
	}
}

// statsHandler returns a handler that serves service statistics.
// TODO: Integrate with stats.Stats to provide real-time statistics
func statsHandler() bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, _ bunrouter.Request) error {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","message":"stats endpoint not yet implemented"}`))
		return nil
	}
}

// healthHandler returns a handler for health checks.
func healthHandler() bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, _ bunrouter.Request) error {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
		return nil
	}
}
