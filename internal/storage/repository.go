package storage

import (
	"chronos-queue/internal/db"
	"context"
	"time"
)

type Repository interface {
	CreateJob(ctx context.Context, arg db.CreateJobParams) (db.Job, error)
	GetJob(ctx context.Context, id string) (db.Job, error)
	ListPendingJobs(ctx context.Context) ([]db.Job, error)
	ClaimJob(ctx context.Context, visibleAfter time.Time) (db.Job, error)
	UpdateJobStatus(ctx context.Context, arg db.UpdateJobStatusParams) error
	GetJobByIdempotencyKey(ctx context.Context, idempotencyKey string) (db.Job, error)
	ExtendVisibility(ctx context.Context, jobID string, visibleAfter time.Time) error
}
