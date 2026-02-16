package worker

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/logger"
	"context"

	"go.uber.org/zap"
)

type JobHandler interface {
	Handle(ctx context.Context, job *pb.Job) error
}

// SimulatedHandler fakes job processing with sleep.
type SimulatedHandler struct{}

func (h *SimulatedHandler) Handle(ctx context.Context, job *pb.Job) error {
	log := logger.Get()
	log.Info("processing job", zap.String("job_id", job.GetId()))
	return nil
}
