package storage

import (
	"chronos-queue/internal/db"
	"context"
)

type Repository interface {
	CreateJob(ctx context.Context, arg db.CreateJobParams) (db.Job, error)
	GetJob(ctx context.Context, id string) (db.Job, error)
	ListPendingJobs(ctx context.Context) ([]db.Job, error)
	ClaimJob(ctx context.Context) (db.Job, error)
	UpdateJobStatus(ctx context.Context, arg db.UpdateJobStatusParams) error
	GetJobByIdempotencyKey(ctx context.Context, idempotencyKey string) (db.Job, error)
}
