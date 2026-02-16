package queue

import (
	"chronos-queue/internal/db"
	"chronos-queue/internal/job"
	"chronos-queue/internal/storage"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type Service struct {
	repo   storage.Repository
	logger *zap.Logger
}

func New(repo storage.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Enqueue(ctx context.Context, jobType string, payload []byte, maxRetries int32, idempotencyKey string) (db.Job, error) {
	params := db.CreateJobParams{
		ID:         uuid.New().String(),
		Type:       jobType,
		Payload:    payload,
		Status:     string(job.StatusPending),
		MaxRetries: maxRetries,
	}

	if idempotencyKey != "" {
		params.IdempotencyKey = pgtype.Text{String: idempotencyKey, Valid: true}
	}

	created, err := s.repo.CreateJob(ctx, params)
	if err != nil {
		s.logger.Error("Failed to enqueue job", zap.Error(err))
		return db.Job{}, err
	}
	s.logger.Info("Enqueued job", zap.String("job_id", created.ID), zap.String("type", created.Type))
	return created, nil
}

func (s *Service) Dequeue(ctx context.Context) (db.Job, error) {
	claimed, err := s.repo.ClaimJob(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.Job{}, nil // No job available, return nil error
		}
		s.logger.Error("Failed to dequeue job", zap.Error(err))
		return db.Job{}, err
	}
	s.logger.Info("Dequeued job", zap.String("job_id", claimed.ID), zap.String("type", claimed.Type))
	return claimed, nil
}

func (s *Service) Complete(ctx context.Context, jobID string) error {
	err := s.repo.UpdateJobStatus(ctx, db.UpdateJobStatusParams{
		ID:     jobID,
		Status: string(job.StatusCompleted),
	})
	if err != nil {
		s.logger.Error("Failed to complete job", zap.String("job_id", jobID), zap.Error(err))
		return err
	}
	s.logger.Info("Completed job", zap.String("job_id", jobID))
	return nil
}

func (s *Service) Fail(ctx context.Context, jobID string) error {
	current, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		s.logger.Error("Failed to get job for failure", zap.String("job_id", jobID), zap.Error(err))
		return err
	}

	var newStatus job.JobStatus
	if current.RetryCount < current.MaxRetries {
		newStatus = job.StatusRetrying
	} else {
		newStatus = job.StatusFailed
	}
	err = s.repo.UpdateJobStatus(ctx, db.UpdateJobStatusParams{
		ID:         jobID,
		Status:     string(newStatus),
		RetryCount: current.RetryCount + 1,
	})

	if err != nil {
		s.logger.Error("Failed to update job status to failed/retrying", zap.String("job_id", jobID), zap.Error(err))
		return err
	}
	s.logger.Info("Updated job status to failed/retrying", zap.String("job_id", jobID), zap.String("new_status", string(newStatus)))
	return nil
}
