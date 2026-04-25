# Chronos Queue

A distributed job queue built with Go, gRPC, and PostgreSQL.

## Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [Docker](https://docs.docker.com/get-docker/)
- [protoc](https://grpc.io/docs/protoc-installation/) (`brew install protobuf`)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) (`brew install sqlc`)
- [k6](https://grafana.com/docs/k6/latest/) (`brew install k6`)

## Getting Started

### 1. Start the database

```bash
docker compose up -d
```

This starts all infrastructure services:

**PostgreSQL** on `localhost:5432`
- Username: `postgres`
- Password: `postgres`
- Database: `chronos`

**Prometheus** on `localhost:9090` — scrapes metrics from the queue service at `host.docker.internal:9091/metrics`

**Grafana** on `localhost:3000` — dashboards with Prometheus and Jaeger pre-configured as datasources (default login: `admin` / `admin`)

**Jaeger** on `localhost:16686` — trace UI, receives OTel traces from all services via OTLP HTTP on port `4318`

### 2. Generate protobuf code

```bash
./bin/generate-proto
```

Generates Go code from `.proto` files into `gen/pb/`. Installs `protoc-gen-go` and `protoc-gen-go-grpc` plugins if missing.

### 3. Run database migrations

```bash
# Apply all migrations
./bin/migrate up

# Check migration status
./bin/migrate status

# Roll back last migration
./bin/migrate down
```

#### Creating a new migration

```bash
./bin/migrate create <name>
```

Example: `./bin/migrate create add_users_table` creates `db/migrations/00002_add_users_table.sql`.

### 4. Generate sqlc code

```bash
./bin/sqlc
```

Reads queries from `db/queries/` and schema from `db/migrations/`, generates type-safe Go code into `internal/db/`.

#### Adding new queries

Write SQL queries in `db/queries/<table>.sql` using sqlc annotations:

```sql
-- name: GetJob :one
SELECT * FROM jobs WHERE id = $1;

-- name: ListPendingJobs :many
SELECT * FROM jobs WHERE status = 'PENDING';

-- name: UpdateJobStatus :exec
UPDATE jobs SET status = $2 WHERE id = $1;
```

Then re-run `./bin/sqlc`.

## Project Structure

```
cmd/
  producer/main.go       # Producer service entry point
  queue/main.go          # Queue service entry point
  worker/main.go         # Worker service entry point
internal/
  config/                # Configuration loading and validation
  db/                    # Generated sqlc code (do not edit)
  job/                   # Job domain model, status, validation
  queue/                 # Queue service logic
  storage/               # Repository interface
  storage/postgres/      # PostgreSQL repository implementation
  worker/                # Worker poll loop and handler
  retry/                 # Backoff and retry policy
proto/                   # Protobuf definitions
gen/pb/                  # Generated protobuf Go code (do not edit)
db/
  migrations/            # Goose migration files
  queries/               # sqlc query files
bin/
  generate-proto         # Generate Go code from proto files
  migrate                # Run database migrations
  sqlc                   # Generate Go code from SQL queries
```

## Database Connection

Default: `postgres://postgres:postgres@localhost:5432/chronos?sslmode=disable`

Override with the `DATABASE_URL` environment variable:

```bash
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" ./bin/migrate up
```

## Observability

### Metrics (Prometheus + Grafana)

The queue service exposes Prometheus metrics at `:9091/metrics`. Prometheus scrapes this every 60s.

Open Grafana at `http://localhost:3000` → add a dashboard → query these metrics:

| Metric | Type | Description |
|---|---|---|
| `chronos_jobs_submitted_total` | Counter | Jobs enqueued |
| `chronos_jobs_completed_total` | Counter | Jobs successfully completed |
| `chronos_jobs_failed_total` | Counter | Jobs that errored |
| `chronos_jobs_retried_total` | Counter | Retry attempts |
| `chronos_queue_depth` | Gauge | Current queue size |
| `chronos_job_duration_seconds` | Histogram | Job processing latency |
| `chronos_worker_pool_active` | Gauge | Busy workers |
| `chronos_worker_pool_idle` | Gauge | Idle workers |

### Traces (OTel + Jaeger)

All three services (`chronos-producer`, `chronos-queue`, `chronos-worker`) send OTel traces to Jaeger via OTLP HTTP.

Set the exporter endpoint before running any service:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

Open the Jaeger UI at `http://localhost:16686` to search and inspect traces. gRPC calls are auto-instrumented — every `SubmitJob`, `DequeueJob`, `CompleteJob`, etc. creates a span automatically.

### Logs (Zap)

Structured logs are written to stdout.

```bash
# development — colored human-readable output (default)
go run ./cmd/queue

# production — JSON structured logs
APP_ENV=production go run ./cmd/queue

# override log level at runtime
LOG_LEVEL=debug go run ./cmd/queue   # debug | info | warn | error
```

## Testing

### Integration Tests

Requires PostgreSQL running. Tests use the default database connection.

```bash
go test ./test/integration/ -v
```

Runs tests for:
- `TestSubmitAndComplete` — submit, dequeue, and complete a job
- `TestFailAndRetry` — fail a job and verify retry behavior
- `TestIdempotency` — duplicate submissions return the same job
- `TestNoJobsAvailable` — dequeue on empty queue
- `TestMultiWorker` — submit 100 jobs, process with 3 concurrent workers, verify no duplicates

### Load Test

Requires the queue service running on `localhost:50051`.

```bash
k6 run test/load/k6.js
```

Ramps up to 10 virtual users submitting jobs via gRPC over 2 minutes. Measures throughput (req/s) and latency (p50/p90/p95).
