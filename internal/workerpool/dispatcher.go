package workerpool

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/logger"
	"context"
	"io"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Dispatcher = the feeder. It's a single goroutine that polls the queue service on a timer ("any jobs available?"), and when it gets
// one, it drops it into the pool's channel. If the channel is full (all workers are busy and the buffer is packed), it stops polling
// until there's room.

type Dispatcher struct {
	pool         *Pool
	queue        pb.WorkerServiceClient
	workerID     string
	pollInterval time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewDispatcher(pool *Pool, queue pb.WorkerServiceClient, workerID string, pollInterval time.Duration) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Dispatcher{
		pool:         pool,
		queue:        queue,
		workerID:     workerID,
		pollInterval: pollInterval,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (d *Dispatcher) Start() {
	log := logger.Get()
	log.Info("dispatcher started", zap.String("workerID", d.workerID))
	ticker := time.NewTicker(d.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			log.Info("dispatcher stopping", zap.String("workerID", d.workerID))
			return
		case <-ticker.C:
			// Backpressure: dont poll if the channel is full
			if len(d.pool.jobs) >= cap(d.pool.jobs) {
				log.Debug("pool full, skipping poll", zap.String("workerID", d.workerID))
				continue
			}
			d.poll(log)
		}
	}
}

func (d *Dispatcher) poll(log *zap.Logger) {
	stream, err := d.queue.PollJob(d.ctx, &pb.WorkerRequest{WorkerId: d.workerID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound { // no jobs available, not an actual error
			return
		}
		log.Error("failed to poll job", zap.String("workerID", d.workerID), zap.Error(err))
		return
	}

	for {
		job, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("stream recv error", zap.String("workerID", d.workerID), zap.Error(err))
			return
		}
		submitted := d.pool.Submit(Job{
			Proto:    job,
			WorkerID: d.workerID,
		})
		if !submitted {
			log.Warn("job received but pool is full, dropping job", zap.String("workerID", d.workerID), zap.String("jobID", job.Id))
			return
		}
	}
}

func (d *Dispatcher) Stop() {
	d.cancel()
}
