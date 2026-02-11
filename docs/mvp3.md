# MVP 3 – Production-Ready

Harden the system for production. DLQ management, leader election, security, graceful shutdown under real conditions, and load testing.

---

## Goal

Make Chronos-Queue **production-grade**: handle every edge case, secure service communication, and prove stability under load.

---

## Prerequisites

- MVP 1 and MVP 2 fully complete and tested

---

## What to Build (in order)

### Phase 1: Dead-Letter Queue (Admin API)

1. `migrations/004_dead_letter_queue.sql` — Add `dlq_reason`, `failed_at` columns to jobs table (or separate DLQ table)
2. Update `internal/queue/service.go` — When a job exceeds `max_retries`, move to `FAILED` with DLQ metadata
3. Implement Admin API RPCs in `proto/admin.proto`:
   - `ListDeadLetterJobs(ListDLQRequest) → ListDLQResponse` — paginated list of failed jobs
   - `RetryJob(RetryJobRequest) → RetryJobResponse` — requeue a DLQ job (reset retry count, set to PENDING)
   - `GetJobDetails(GetJobRequest) → JobDetailResponse` — full job details with retry history
4. Wire Admin gRPC server (can run on the queue service or as a separate port)

### Phase 2: Leader Election

5. `internal/scheduler/election.go` — Leader election for singleton tasks
   - Use PostgreSQL advisory locks (`pg_try_advisory_lock`) or Redis Redlock
   - Only the leader runs: stale job reaper, DLQ mover, rebalancing
6. `internal/scheduler/lease.go` — Lease renewal
   - Leader periodically renews its lease
   - If lease expires, another instance takes over
7. `internal/scheduler/rebalance.go` — Job rebalancing (optional)
   - Detect workers with disproportionate load
   - Reassign unclaimed jobs

### Phase 3: Stale Job Reaper

8. Add a background goroutine (leader-only) that periodically scans for:
   - Jobs stuck in `IN_PROGRESS` past visibility timeout with no heartbeat
   - Move them back to `PENDING` or `RETRYING`
9. Log and emit metrics for reaped jobs

### Phase 4: Graceful Shutdown (Full)

10. Enhance `internal/workerpool/shutdown.go`:
    - Listen for `SIGTERM` and `SIGINT`
    - Stop accepting new jobs immediately
    - Drain in-flight jobs with a configurable deadline
    - If deadline exceeded, log and force-stop remaining jobs
    - Release any held locks/leases
11. Ensure gRPC servers drain connections properly
12. Test: send SIGTERM during processing, verify no job loss

### Phase 5: Security

13. `internal/auth/api_key.go` — API key authentication for producers
    - gRPC interceptor validates API key from metadata
    - Keys loaded from config/env
14. `internal/auth/mtls.go` — mTLS between services
    - TLS certificate loading
    - Mutual authentication for queue ↔ worker communication
15. Generate test certs in `scripts/generate-certs.sh`
16. Rate limiting enforcement via Redis (from MVP 2) with proper error responses

### Phase 6: Job Execution Timeout

17. Enforce per-job execution timeout in the worker
    - Use `context.WithTimeout` based on job's `execution_timeout` field
    - Cancel job if timeout exceeded
    - Report as failure (counts toward retry)
18. `migrations/005_add_execution_timeout.sql` — Add `execution_timeout` column if not already present

### Phase 7: Retry History

19. `migrations/006_add_retry_history.sql` — Create `job_retry_history` table
    - `job_id`, `attempt`, `error_message`, `worker_id`, `attempted_at`
20. Record each retry attempt for debugging
21. Expose via `GetJobDetails` admin RPC

### Phase 8: Load Testing & Benchmarks

22. `test/load/k6.js` — Comprehensive load test scenarios:
    - Sustained throughput: 1000 jobs/sec for 5 minutes
    - Burst: 10,000 jobs in 10 seconds
    - Failure rate: 30% job failure rate, verify retry behavior
    - Worker crash: kill workers during processing, verify recovery
23. `scripts/load-test.sh` — Orchestrate load tests with reporting
24. Document results: throughput, p50/p95/p99 latency, memory usage, recovery time

### Phase 9: Deployment

25. `deploy/docker/producer.Dockerfile` — Multi-stage build
26. `deploy/docker/queue.Dockerfile`
27. `deploy/docker/worker.Dockerfile`
28. Update `docker-compose.yml` with production-like configuration:
    - Health checks
    - Resource limits
    - Proper networking
    - Environment variable management
29. `scripts/run-local.sh` — One-command local setup
30. `scripts/migrate.sh` — Run migrations safely

### Phase 10: Documentation & Interview Prep

31. Update `README.md`:
    - Architecture diagram
    - Quick start guide
    - Configuration reference
    - API reference
32. Document design decisions:
    - Why at-least-once delivery
    - Why `SKIP LOCKED` over distributed locks for job claiming
    - How visibility timeouts prevent job loss
    - Goroutine lifecycle and leak prevention
    - What breaks under extreme load and how to mitigate

---

## Files Touched

```
internal/queue/service.go (updated)
internal/scheduler/election.go
internal/scheduler/lease.go
internal/scheduler/rebalance.go
internal/workerpool/shutdown.go (updated)
internal/auth/api_key.go
internal/auth/mtls.go
proto/admin.proto (updated)
migrations/004_dead_letter_queue.sql
migrations/005_add_execution_timeout.sql
migrations/006_add_retry_history.sql
deploy/docker/producer.Dockerfile
deploy/docker/queue.Dockerfile
deploy/docker/worker.Dockerfile
docker-compose.yml (updated)
scripts/generate-certs.sh
scripts/load-test.sh
scripts/run-local.sh
scripts/migrate.sh
test/load/k6.js (updated)
README.md
```

---

## What "Done" Looks Like

- [ ] DLQ jobs queryable and re-queueable via Admin API
- [ ] Leader election prevents duplicate singleton tasks
- [ ] Stale jobs automatically reaped and retried
- [ ] SIGTERM triggers clean shutdown with no job loss
- [ ] mTLS secures service-to-service communication
- [ ] API keys authenticate producers
- [ ] Per-job execution timeout enforced
- [ ] Retry history recorded and queryable
- [ ] Load test proves linear scaling and recovery from failures
- [ ] Dockerfiles build and run cleanly
- [ ] README documents architecture, setup, and design decisions

---

## Success Metrics (from PRD)

- Zero job loss during worker crashes
- Stable memory usage under sustained load
- Linear throughput scaling with workers
- Predictable retry behavior with backoff + jitter
