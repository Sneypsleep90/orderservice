package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ordersProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_processed_total",
		Help: "Total number of processed orders",
	})

	ordersProcessErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_process_errors_total",
		Help: "Total number of errors during order processing",
	})

	orderProcessDurationSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "order_process_duration_seconds",
		Help:    "Time spent processing orders",
		Buckets: prometheus.DefBuckets,
	})
)
