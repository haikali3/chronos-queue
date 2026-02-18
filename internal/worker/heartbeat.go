package worker

import (
	"chronos-queue/internal/logger"
	"chronos-queue/internal/queue"
	"context"
	"time"

	"go.uber.org/zap"
)

type Heartbeat struct {
	jobID        string
	interval     time.Duration
	queueService *queue.Service
	cancel       context.CancelFunc
	done         chan struct{}
}

func NewHeartbeat(jobID string, interval time.Duration, queueService *queue.Service) *Heartbeat {
	return &Heartbeat{
		jobID:        jobID,
		interval:     interval,
		queueService: queueService,
		done:         make(chan struct{}),
	}
}

func (h *Heartbeat) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	go func() {
		defer close(h.done)
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := h.queueService.ExtendVisibility(ctx, h.jobID); err != nil {
					logger.Get().Warn("failed to extend job visibility", zap.String("job_id", h.jobID), zap.Error(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (h *Heartbeat) Stop() {
	if h.cancel != nil {
		h.cancel()
		<-h.done
	}
}
