package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_server_requests_total",
		Help: "Total gRPC requests",
	}, []string{"service", "method", "code", "status"}) // 4个标签

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grpc_server_request_duration_seconds",
		Help:    "gRPC request duration",
		Buckets: []float64{0.1, 0.3, 0.5, 1.0, 2.5, 5.0},
	}, []string{"service", "method", "code"}) // 3个标签
)
