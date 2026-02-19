# MVP 2 – Concurrency & Scale

Scale the system horizontally. Multiple workers, goroutine pools, Redis coordination, and observability.

---

## Goal

Go from a single-threaded worker to a **concurrent, multi-worker system** with proper coordination, backpressure, and visibility into what's happening.

---

## Prerequisites

- MVP 1 fully complete and tested

---

## What to Build (in order)

### Phase 1: Worker Pool

<!-- 1. `internal/workerpool/pool.go` — Worker pool with configurable size (N goroutines)
   - Accept jobs via a channel
   - Each goroutine pulls from the channel and processes
   - Track active/idle worker counts
2. `internal/workerpool/dispatcher.go` — Dispatcher goroutine
   - Polls queue service for jobs
   - Sends jobs into the worker pool channel
   - Implements backpressure: stop polling when channel is full
3. `internal/workerpool/shutdown.go` — Graceful shutdown
   - Stop accepting new jobs
   - Wait for in-flight jobs to complete (with timeout)
   - Use `sync.WaitGroup` + `context.Context`
4. Update `cmd/worker/main.go` to use the pool instead of a single poll loop -->

### Phase 2: Visibility Timeout

<!-- 5. `internal/queue/visibility.go` — Visibility timeout logic
   - When a job is claimed, set `visible_after = now + timeout`
   - Jobs past their visibility timeout become re-claimable
   - Prevents stuck jobs from blocking the queue
6. `migrations/002_add_visibility_timeout.sql` — Add `visible_after` column to jobs table
7. Update `ClaimJob` query: only claim jobs where `visible_after IS NULL OR visible_after <= now` -->

### Phase 3: Heartbeat

<!-- 8. `internal/worker/heartbeat.go` — Workers send periodic heartbeats for in-progress jobs
   - Extends `visible_after` on each heartbeat
   - If worker crashes, visibility timeout expires naturally
9. Add heartbeat RPC to `proto/worker.proto`: `Heartbeat(HeartbeatRequest) → Ack` -->

### Phase 4: Redis Integration

<!-- 10. `internal/storage/redis/lock.go` — Distributed locking
    - Used for coordinating job claims across multiple queue service instances
    - Simple lock/unlock with TTL
11. `internal/storage/redis/limiter.go` — Rate limiting
    - Token bucket or sliding window on job submission
    - Configurable per job type
12. Update `docker-compose.yml` to include Redis
13. Update `internal/config/config.go` with Redis connection config -->

### Phase 5: Multiple Workers & Horizontal Scaling

<!-- 14. Ensure multiple worker instances can run simultaneously
    - Each worker registers with a unique worker ID
    - `SELECT ... FOR UPDATE SKIP LOCKED` prevents double-claiming -->
15. Test with 3+ worker instances via docker-compose scale
<!-- 16. Add worker ID to job tracking (`claimed_by` column)
17. `migrations/003_add_worker_tracking.sql` -->

### Phase 6: Observability

<!-- 18. `internal/logger/logger.go` — Structured logging with zerolog
    - Request IDs
    - Job IDs in all log lines
    - Log levels: debug, info, warn, error
19. `internal/observability/metrics.go` — Prometheus metrics
    - `chronos_jobs_submitted_total` (counter)
    - `chronos_jobs_completed_total` (counter)
    - `chronos_jobs_failed_total` (counter)
    - `chronos_jobs_retried_total` (counter)
    - `chronos_queue_depth` (gauge)
    - `chronos_job_duration_seconds` (histogram)
    - `chronos_worker_pool_active` (gauge)
    - `chronos_worker_pool_idle` (gauge) -->
20. `internal/observability/tracing.go` — OpenTelemetry tracing
    - Trace spans: job submission → queue → worker execution
    - Propagate trace context through gRPC metadata
<!-- 21. Expose `/metrics` endpoint on each service -->
<!-- 22. Add Prometheus + Grafana to `docker-compose.yml` (optional) -->

### Phase 7: Backpressure & Flow Control

23. `internal/workerpool/metrics.go` — Pool utilization metrics
    - Expose channel buffer fill percentage
    - Log warnings when pool is saturated
24. Implement adaptive polling: slow down poll frequency when pool utilization is high

### Phase 8: Testing at Scale

25. `test/integration/worker_test.go` — Multi-worker integration tests
    - Submit 100 jobs, verify all completed with 3 workers
    - Verify no duplicate processing
    - Verify metrics counts match
26. `test/load/k6.js` — Basic load test script
    - Ramp up job submission rate
    - Measure throughput and latency

---

## Files Touched

```
internal/workerpool/pool.go
internal/workerpool/dispatcher.go
internal/workerpool/shutdown.go
internal/workerpool/metrics.go
internal/queue/visibility.go
internal/worker/heartbeat.go
internal/storage/redis/lock.go
internal/storage/redis/limiter.go
internal/observability/logger.go
internal/observability/metrics.go
internal/observability/tracing.go
internal/config/config.go (updated)
cmd/worker/main.go (updated)
proto/worker.proto (updated)
migrations/002_add_visibility_timeout.sql
migrations/003_add_worker_tracking.sql
docker-compose.yml (updated)
test/integration/worker_test.go
test/load/k6.js
```

---

## What "Done" Looks Like

- [ ] Worker processes N jobs concurrently via goroutine pool
- [ ] Backpressure stops polling when pool is saturated
- [ ] Visibility timeout reclaims stuck jobs
- [ ] Heartbeats extend visibility for long-running jobs
- [ ] Multiple worker instances run without conflicts
- [ ] Redis rate limiting throttles job submission
- [ ] Prometheus metrics exposed and accurate
- [ ] Structured logs with job/request context
- [ ] Load test shows linear throughput scaling with workers

---

## What We're NOT Doing Yet

- No DLQ admin API (MVP 3)
- No leader election (MVP 3)
- No mTLS / API key auth (MVP 3)
- No k8s deployment (future)
