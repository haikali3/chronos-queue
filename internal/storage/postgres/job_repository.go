package postgres

import (
	"chronos-queue/internal/db"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepository struct {
	queries *db.Queries
}

func NewJobRepository(pool *pgxpool.Pool) *JobRepository {
	return &JobRepository{queries: db.New(pool)}
}

func (r *JobRepository) CreateJob(ctx context.Context, arg *db.CreateJobParams) (db.Job, error) {
	return r.queries.CreateJob(ctx, *arg)
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

func (r *JobRepository) UpdateJobStatus(ctx context.Context, arg *db.UpdateJobStatusParams) error {
	return r.queries.UpdateJobStatus(ctx, *arg)
}
