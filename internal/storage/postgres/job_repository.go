package postgres

import (
	"chronos-queue/internal/db"
	"chronos-queue/internal/job"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

func (r *JobRepository) ClaimJob(ctx context.Context, visibleAfter time.Time) (db.Job, error) {
	return r.queries.ClaimJob(ctx, pgtype.Timestamptz{Time: visibleAfter, Valid: true})
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

func (r *JobRepository) GetJobByIdempotencyKey(ctx context.Context, idempotencyKey string) (db.Job, error) {
	return r.queries.GetJobByIdempotencyKey(ctx, pgtype.Text{String: idempotencyKey, Valid: true})
}

func (r *JobRepository) ExtendVisibility(ctx context.Context, jobID string, visibleAfter time.Time) error {
	return r.queries.UpdateJobVisibility(ctx, db.UpdateJobVisibilityParams{
		ID:           jobID,
		VisibleAfter: pgtype.Timestamptz{Time: visibleAfter, Valid: true},
	})
}
