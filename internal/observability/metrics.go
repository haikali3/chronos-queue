package observability

import "github.com/prometheus/client_golang/prometheus"

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
