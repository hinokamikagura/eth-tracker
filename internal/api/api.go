package api

import (
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/uptrace/bunrouter"
)

func New() *bunrouter.Router {
	router := bunrouter.New()

	router.GET("/metrics", metricsHandler())

	return router
}

func metricsHandler() bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, _ bunrouter.Request) error {
		metrics.WritePrometheus(w, true)
		return nil
	}
}
