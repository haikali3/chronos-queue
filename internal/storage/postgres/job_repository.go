package postgres

import (
	"chronos-queue/internal/db"
	"chronos-queue/internal/job"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepository struct {
	queries *db.Queries
}

func NewJobRepository(pool *pgxpool.Pool) *JobRepository {
	return &JobRepository{queries: db.New(pool)}
}

func (r *JobRepository) CreateJob(ctx context.Context, arg db.CreateJobParams) (db.Job, error) {
	return r.queries.CreateJob(ctx, arg)
}

func (r *JobRepository) GetJob(ctx context.Context, id string) (db.Job, error) {
	return r.queries.GetJob(ctx, id)
}

func (r *JobRepository) ListPendingJobs(ctx context.Context) ([]db.Job, error) {
	return r.queries.ListPendingJobs(ctx)
}

func (r *JobRepository) ClaimJob(ctx context.Context) (db.Job, error) {
	return r.queries.ClaimJob(ctx)
}

func (r *JobRepository) UpdateJobStatus(ctx context.Context, arg db.UpdateJobStatusParams) error {
	currentJob, err := r.queries.GetJob(ctx, arg.ID)
	if err != nil {
		return err
	}

	from := job.JobStatus(currentJob.Status)
	to := job.JobStatus(arg.Status)
	if !job.IsValidTransition(from, to) {
		return fmt.Errorf("invalid job status transition: %s -> %s", from, to)
	}

	return r.queries.UpdateJobStatus(ctx, arg)
}
