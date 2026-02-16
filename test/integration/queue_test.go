package integration

import (
	"chronos-queue/internal/job"
	"chronos-queue/internal/queue"
	"chronos-queue/internal/storage/postgres"
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func setupService(t *testing.T) (*queue.Service, *pgxpool.Pool) {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/chronos?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Clean jobs table before each test.
	_, err = pool.Exec(context.Background(), "DELETE FROM jobs")
	if err != nil {
		t.Fatalf("failed to clean jobs table: %v", err)
	}

	repo := postgres.NewJobRepository(pool)
	logger, _ := zap.NewDevelopment()
	svc := queue.New(repo, logger)

	return svc, pool
}

func TestSubmitAndComplete(t *testing.T) {
	svc, pool := setupService(t)
	defer pool.Close()
	ctx := context.Background()

	// Submit a job.
	created, err := svc.Enqueue(ctx, "email", []byte(`{"to":"test@example.com"}`), 3, "")
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}
	if created.Status != string(job.StatusPending) {
		t.Fatalf("expected PENDING, got %s", created.Status)
	}

	// Worker dequeues it.
	claimed, err := svc.Dequeue(ctx)
	if err != nil {
		t.Fatalf("failed to dequeue: %v", err)
	}
	if claimed.ID != created.ID {
		t.Fatalf("expected job %s, got %s", created.ID, claimed.ID)
	}

	// Worker completes it.
	err = svc.Complete(ctx, claimed.ID)
	if err != nil {
		t.Fatalf("failed to complete: %v", err)
	}

	// Verify final state.
	var status string
	err = pool.QueryRow(ctx, "SELECT status FROM jobs WHERE id = $1", claimed.ID).Scan(&status)
	if err != nil {
		t.Fatalf("failed to query job: %v", err)
	}
	if status != string(job.StatusCompleted) {
		t.Fatalf("expected COMPLETED, got %s", status)
	}
}

func TestFailAndRetry(t *testing.T) {
	svc, pool := setupService(t)
	defer pool.Close()
	ctx := context.Background()

	// Submit a job with max 2 retries.
	created, err := svc.Enqueue(ctx, "email", []byte(`{"to":"test@example.com"}`), 2, "")
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	// Dequeue and fail it — should become RETRYING.
	_, err = svc.Dequeue(ctx)
	if err != nil {
		t.Fatalf("failed to dequeue: %v", err)
	}
	err = svc.Fail(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to fail job: %v", err)
	}

	var status string
	var retryCount int32
	err = pool.QueryRow(ctx, "SELECT status, retry_count FROM jobs WHERE id = $1", created.ID).Scan(&status, &retryCount)
	if err != nil {
		t.Fatalf("failed to query job: %v", err)
	}
	if status != string(job.StatusRetrying) {
		t.Fatalf("expected RETRYING, got %s", status)
	}
	if retryCount != 1 {
		t.Fatalf("expected retry_count 1, got %d", retryCount)
	}
}

func TestIdempotency(t *testing.T) {
	svc, pool := setupService(t)
	defer pool.Close()
	ctx := context.Background()

	// Submit with idempotency key.
	first, err := svc.Enqueue(ctx, "email", []byte(`{"to":"test@example.com"}`), 3, "unique-key-123")
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	// Submit again with same key — should return existing job.
	second, err := svc.Enqueue(ctx, "email", []byte(`{"to":"test@example.com"}`), 3, "unique-key-123")
	if err != nil {
		t.Fatalf("failed to enqueue duplicate: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected same job ID, got %s and %s", first.ID, second.ID)
	}
}

func TestNoJobsAvailable(t *testing.T) {
	svc, pool := setupService(t)
	defer pool.Close()
	ctx := context.Background()

	// Dequeue with no jobs — should return ErrJobNotFound.
	_, err := svc.Dequeue(ctx)
	if err != queue.ErrJobNotFound {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}
