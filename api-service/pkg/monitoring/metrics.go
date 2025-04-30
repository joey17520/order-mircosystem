package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"service", "method", "path", "status"})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: []float64{0.1, 0.3, 0.5, 1.0, 2.5, 5.0},
	}, []string{"service", "method", "path"})

	ErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "errors_total",
		Help: "Total errors",
	}, []string{"service", "type"})
)
