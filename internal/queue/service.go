package queue

import (
	"chronos-queue/internal/db"
	"chronos-queue/internal/job"
	"chronos-queue/internal/requestid"
	"chronos-queue/internal/retry"
	"chronos-queue/internal/storage"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("chronos-queue")

type Service struct {
	repo          storage.Repository
	logger        *zap.Logger
	visibilityCfg VisibilityConfig
}

func New(repo storage.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
		visibilityCfg: VisibilityConfig{
			timeout: 30 * time.Second,
			// In a real implementation, this could be configurable per job type or even per job
		},
	}
}

func (s *Service) Enqueue(ctx context.Context, jobType string, payload []byte, maxRetries int32, idempotencyKey string) (db.Job, error) {
	ctx, span := tracer.Start(ctx, "Service.Enqueue")
	defer span.End()

	jobID := uuid.New().String()
	span.SetAttributes(
		attribute.String("job_type", jobType),
		attribute.String("job_id", jobID),
	)

	requestID, _ := requestid.FromContext(ctx)
	if idempotencyKey != "" {
		existing, err := s.repo.GetJobByIdempotencyKey(ctx, idempotencyKey)
		if err == nil {
			s.logger.Info("duplicate idempotency key, returning existing job", zap.String("job_id", existing.ID))
			return existing, nil
		}
	}

	params := db.CreateJobParams{
		ID:         jobID,
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
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())

		s.logger.Error("Failed to enqueue job", zap.String("request_id", requestID), zap.Error(err))
		return db.Job{}, err
	}
	s.logger.Info("Enqueued job", zap.String("request_id", requestID), zap.String("job_id", created.ID), zap.String("type", created.Type))
	return created, nil
}

func (s *Service) Dequeue(ctx context.Context, workerID string) (db.Job, error) {
	ctx, span := tracer.Start(ctx, "Service.Dequeue")
	defer span.End()

	span.SetAttributes(
		attribute.String("worker_id", workerID),
	)

	requestID, _ := requestid.FromContext(ctx)
	claimed, err := s.repo.ClaimJob(ctx, time.Now().Add(s.visibilityCfg.timeout), workerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {

			s.logger.Info("No jobs available to dequeue", zap.String("request_id", requestID))
			return db.Job{}, ErrJobNotFound
		}
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		s.logger.Error("Failed to dequeue job", zap.String("request_id", requestID), zap.Error(err))
		return db.Job{}, err
	}
	s.logger.Info("Dequeued job", zap.String("request_id", requestID), zap.String("job_id", claimed.ID), zap.String("type", claimed.Type))
	return claimed, nil
}

func (s *Service) Complete(ctx context.Context, jobID string) error {
	ctx, span := tracer.Start(ctx, "Service.Complete")
	span.SetAttributes(
		attribute.String("job_id", jobID),
	)
	defer span.End()
	requestID, _ := requestid.FromContext(ctx)
	_, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, "job not found")

			s.logger.Error("Job not found for completion", zap.String("request_id", requestID), zap.String("job_id", jobID))
			return ErrJobNotFound
		}
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())

		s.logger.Error("Failed to get job for completion", zap.String("request_id", requestID), zap.String("job_id", jobID), zap.Error(err))
		return err
	}

	err = s.repo.UpdateJobStatus(ctx, db.UpdateJobStatusParams{
		ID:     jobID,
		Status: string(job.StatusCompleted),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())

		s.logger.Error("Failed to complete job", zap.String("request_id", requestID), zap.String("job_id", jobID), zap.Error(err))
		return err
	}
	s.logger.Info("Completed job", zap.String("request_id", requestID), zap.String("job_id", jobID))
	return nil
}

func (s *Service) Fail(ctx context.Context, jobID string) error {
	ctx, span := tracer.Start(ctx, "Service.Fail")
	defer span.End()
	span.SetAttributes(
		attribute.String("job_id", jobID),
	)
	requestID, _ := requestid.FromContext(ctx)
	current, err := s.repo.GetJob(ctx, jobID)
	if errors.Is(err, pgx.ErrNoRows) {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, "job not found")

		s.logger.Error("Job not found for failure", zap.String("request_id", requestID), zap.String("job_id", jobID))
		return ErrJobNotFound
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())

		s.logger.Error("Failed to get job for failure", zap.String("request_id", requestID), zap.String("job_id", jobID), zap.Error(err))
		return err
	}

	var newStatus job.JobStatus
	var retryCount int32
	var nextRetry pgtype.Timestamptz

	if current.RetryCount < current.MaxRetries {
		newStatus = job.StatusRetrying
		retryCount = current.RetryCount + 1
		nextRetry = pgtype.Timestamptz{
			Time:  retry.NextRetryAt(retryCount),
			Valid: true,
		}
	} else {
		newStatus = job.StatusFailed
		retryCount = current.RetryCount
	}

	err = s.repo.UpdateJobStatus(ctx, db.UpdateJobStatusParams{
		ID:          jobID,
		Status:      string(newStatus),
		RetryCount:  retryCount,
		NextRetryAt: nextRetry,
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		s.logger.Error("Failed to update job status to failed/retrying", zap.String("request_id", requestID), zap.String("job_id", jobID), zap.Error(err))
		return err
	}

	s.logger.Info("Updated job status to failed/retrying", zap.String("request_id", requestID), zap.String("job_id", jobID), zap.String("new_status", string(newStatus)))
	return nil
}
