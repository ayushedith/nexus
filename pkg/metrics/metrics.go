package metrics

import (
    "net/http"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    RequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "nexus_requests_total", Help: "Total requests executed."},
        []string{"status"},
    )
)

func init() {
    prometheus.MustRegister(RequestsTotal)
}

// Handler returns an HTTP handler that serves Prometheus metrics.
func Handler() http.Handler {
    return promhttp.Handler()
}
