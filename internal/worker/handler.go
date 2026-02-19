package worker

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/logger"
	"chronos-queue/internal/observability"
	"context"

	"go.uber.org/zap"
)

type JobHandler interface {
	Handle(ctx context.Context, job *pb.Job) error
}

// SimulatedHandler fakes job processing with sleep.
type SimulatedHandler struct {
	metrics *observability.Metrics
}

func NewSimulatedHandler(metrics *observability.Metrics) *SimulatedHandler {
	return &SimulatedHandler{metrics: metrics}
}

func (h *SimulatedHandler) Handle(ctx context.Context, job *pb.Job) error {
	log := logger.Get()
	log.Info("processing job", zap.String("job_id", job.GetId()))

	h.metrics.JobsCompleted.Inc()

	return nil
}
