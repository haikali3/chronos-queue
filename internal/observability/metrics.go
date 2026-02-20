package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	JobsSubmitted    prometheus.Counter
	JobsCompleted    prometheus.Counter
	JobsFailed       prometheus.Counter
	JobsRetried      prometheus.Counter
	QueueDepth       prometheus.Gauge
	JobDuration      prometheus.Histogram
	WorkerPoolActive prometheus.Gauge
	WorkerPoolIdle   prometheus.Gauge
}

func NewMetrics() *Metrics {
	m := &Metrics{
		JobsSubmitted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "chronos_jobs_submitted_total",
			Help: "Total number of jobs submitted to the queue",
		}),
		JobsCompleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "chronos_jobs_completed_total",
			Help: "Total number of jobs successfully completed",
		}),
		JobsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "chronos_jobs_failed_total",
			Help: "Total number of jobs that failed",
		}),
		JobsRetried: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "chronos_jobs_retried_total",
			Help: "Total number of job retries",
		}),
		QueueDepth: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "chronos_queue_depth",
			Help: "Current number of jobs in the queue",
		}),
		JobDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "chronos_job_duration_seconds",
			Help:    "Duration of job processing in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		WorkerPoolActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "chronos_worker_pool_active",
			Help: "Number of active workers in the pool",
		}),
		WorkerPoolIdle: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "chronos_worker_pool_idle",
			Help: "Number of idle workers in the pool",
		}),
	}
	prometheus.MustRegister(
		m.JobsSubmitted,
		m.JobsCompleted,
		m.JobsFailed,
		m.JobsRetried,
		m.QueueDepth,
		m.JobDuration,
		m.WorkerPoolActive,
		m.WorkerPoolIdle,
	)
	return m
}
