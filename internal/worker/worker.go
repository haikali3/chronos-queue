package worker

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

type Worker struct {
	id           string
	queue        pb.WorkerServiceClient
	handler      JobHandler
	pollInterval time.Duration
}

func New(id string, queue pb.WorkerServiceClient, handler JobHandler, pollInterval time.Duration) *Worker {
	return &Worker{
		id:           id,
		queue:        queue,
		handler:      handler,
		pollInterval: pollInterval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	log := logger.Get()
	log.Info("worker started", zap.String("worker_id", w.id))

	for {
		select {
		case <-ctx.Done():
			log.Info("worker stopping", zap.String("worker_id", w.id))
			return
		default:
			w.poll(ctx)
			time.Sleep(w.pollInterval)
		}
	}
}

func (w *Worker) poll(ctx context.Context) {
	log := logger.Get()
	stream, err := w.queue.PollJob(ctx, &pb.WorkerRequest{WorkerId: w.id})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return // no jobs, silent
		}
		log.Error("failed to dequeue job", zap.Error(err))
		return
	}

	for {
		job, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				//no jobs available: normal
				return
			}
			log.Error("stream error", zap.Error(err))
			return
		}

		success := true
		handleErr := w.handler.Handle(ctx, job)
		if handleErr != nil {
			log.Error("job handling failed", zap.String("job_id", job.GetId()), zap.Error(handleErr))
			success = false
		}

		_, err = w.queue.ReportResult(ctx, &pb.JobResult{
			JobId:    job.GetId(),
			WorkerId: w.id,
			Success:  success,
		})
		if err != nil {
			log.Error("failed to report job result", zap.String("job_id", job.GetId()), zap.Error(err))
		}
	}
}
