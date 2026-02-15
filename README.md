# Chronos Queue

A distributed job queue built with Go, gRPC, and PostgreSQL.

## Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [Docker](https://docs.docker.com/get-docker/)
- [protoc](https://grpc.io/docs/protoc-installation/) (`brew install protobuf`)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) (`brew install sqlc`)

## Getting Started

### 1. Start the database

```bash
docker compose up -d
```

This starts PostgreSQL on `localhost:5432` with:
- Username: `postgres`
- Password: `postgres`
- Database: `chronos`

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
