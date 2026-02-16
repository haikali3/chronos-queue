# MVP 1 – Core Queue

The foundation. Get a single job submitted, persisted, picked up by a worker, and completed — with basic retry on failure.

---

## Goal

A working end-to-end job queue: **submit → persist → process → complete/retry**.

No concurrency pools, no Redis, no observability yet. Just correctness.

---

## What to Build (in order)

### Phase 1: Project Skeleton & Protobuf

<!-- 1. Initialize Go module (`go mod init github.com/haikaltahar/chronos-queue`)
2. Set up `cmd/producer/main.go`, `cmd/queue/main.go`, `cmd/worker/main.go` as empty entry points
3. Define protobuf contracts in `proto/`:
   - `common.proto` — shared types: `Job`, `JobStatus` enum (`PENDING`, `IN_PROGRESS`, `COMPLETED`, `RETRYING`, `FAILED`)
   - `producer.proto` — `SubmitJob(JobRequest) → JobResponse`
   - `worker.proto` — `PollJobs(WorkerRequest) → stream Job`, `ReportResult(JobResult) → Ack`
   - `admin.proto` — `ListJobs`, `RetryJob` (stub for now)
4. Generate Go code from protos -->

### Phase 2: Job Domain Model

<!-- 5. `internal/job/job.go` — Job struct with fields: `ID`, `Type`, `Payload`, `Status`, `RetryCount`, `MaxRetries`, `IdempotencyKey`, `CreatedAt`, `UpdatedAt`
6. `internal/job/status.go` — State machine: valid transitions (`PENDING→IN_PROGRESS`, `IN_PROGRESS→COMPLETED`, `IN_PROGRESS→RETRYING`, `RETRYING→IN_PROGRESS`, `RETRYING→FAILED`)
7. `internal/job/validator.go` — Validate job payloads (type not empty, max retries >= 0, valid JSON payload) -->

### Phase 3: PostgreSQL Storage

<!-- 8. `migrations/001_create_jobs_table.sql` — Create `jobs` table matching the data model -->
<!-- 9. `internal/storage/repository.go` — Repository interface: `CreateJob`, `GetJob`, `ListPendingJobs`, `ClaimJob`, `UpdateJobStatus`
10. `internal/storage/postgres/job_repository.go` — PostgreSQL implementation
    - `ClaimJob`: use `SELECT ... FOR UPDATE SKIP LOCKED` to safely claim a pending job
    - `UpdateJobStatus`: enforce valid state transitions -->

### Phase 4: Configuration

<!-- 11. `internal/config/config.go` — Load config from env vars: DB connection string, gRPC ports, worker poll interval
12. `internal/config/validation.go` — Validate required config on startup -->

### Phase 5: Queue Service

<!-- 13. `internal/queue/service.go` — Core queue logic: `Enqueue(job)`, `Dequeue() → job`, `Complete(jobID)`, `Fail(jobID)`
14. `internal/queue/errors.go` — Domain errors: `ErrJobNotFound`, `ErrInvalidTransition`, `ErrDuplicateIdempotencyKey`
15. Wire up gRPC server in `cmd/queue/main.go` — expose producer and worker RPCs -->

### Phase 6: Producer Service

<!-- 16. `cmd/producer/main.go` — gRPC server that accepts `SubmitJob`, validates input, calls Queue Service to enqueue
17. Keep it thin — validation + forward to queue -->

### Phase 7: Single Worker

<!-- 18. `internal/worker/worker.go` — Simple poll loop:
    - Poll queue service for a job via gRPC
    - Execute job (simulated — sleep + log for now)
    - Report success or failure back
19. `internal/worker/handler.go` — Job handler interface + a default simulated handler -->
<!-- 20. Wire up in `cmd/worker/main.go` -->

### Phase 8: Basic Retry Logic

<!-- 21. `internal/retry/backoff.go` — Exponential backoff: `delay = base * 2^attempt`
22. `internal/retry/policy.go` — Retry policy: check `retry_count < max_retries`, compute next retry delay
23. Integrate into worker: on failure, update job to `RETRYING` with next retry timestamp
24. Queue service: only dequeue jobs where `status = PENDING` OR (`status = RETRYING` AND `next_retry_at <= now`) -->

### Phase 9: Idempotency

25. `internal/job/idempotency.go` — Check idempotency key before creating a job (unique constraint on `idempotency_key` in DB)
26. Return existing job if duplicate key submitted

### Phase 10: Docker Compose & Integration

27. `docker-compose.yml` — PostgreSQL + all three services
28. `Makefile` — targets: `proto`, `build`, `run`, `migrate`, `test`
29. `test/integration/queue_test.go` — End-to-end test: submit job → verify it gets processed → verify final state

---

## Files Touched

```
cmd/producer/main.go
cmd/queue/main.go
cmd/worker/main.go
internal/config/config.go
internal/config/validation.go
internal/job/job.go
internal/job/status.go
internal/job/validator.go
internal/job/idempotency.go
internal/queue/service.go
internal/queue/errors.go
internal/worker/worker.go
internal/worker/handler.go
internal/retry/backoff.go
internal/retry/policy.go
internal/storage/repository.go
internal/storage/postgres/job_repository.go
proto/common.proto
proto/producer.proto
proto/worker.proto
proto/admin.proto
migrations/001_create_jobs_table.sql
docker-compose.yml
Makefile
test/integration/queue_test.go
```

---

## What "Done" Looks Like

- [ ] `SubmitJob` RPC accepts a job and persists it to PostgreSQL
- [ ] Worker polls and picks up pending jobs
- [ ] Successful jobs move to `COMPLETED`
- [ ] Failed jobs retry with exponential backoff up to `max_retries`
- [ ] Jobs exceeding retries move to `FAILED`
- [ ] Duplicate idempotency keys return existing job
- [ ] `docker-compose up` starts the full system
- [ ] Integration test passes end-to-end

---

## What We're NOT Doing Yet

- No worker pools / multiple goroutines (MVP 2)
- No Redis (MVP 2)
- No metrics or tracing (MVP 2)
- No DLQ queryable API (MVP 3)
- No leader election (MVP 3)
- No graceful shutdown (MVP 3)
- No mTLS or API key auth (MVP 3)
