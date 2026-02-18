package worker

import (
	"chronos-queue/internal/queue"
	"context"
	"time"
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

	go func() {
		h.cancel = cancel
		defer close(h.done)
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.queueService.ExtendVisibility(ctx, h.jobID)
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
