package grpc

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/observability"
	"chronos-queue/internal/queue"
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WorkerHandler struct {
	pb.UnimplementedWorkerServiceServer
	svc     *queue.Service
	metrics *observability.Metrics // prometheus,opentelemetry metrics and jaegar metrics
}

func NewWorkerHandler(svc *queue.Service, metrics *observability.Metrics) *WorkerHandler {
	return &WorkerHandler{svc: svc, metrics: metrics}
}

func (h *WorkerHandler) PollJob(req *pb.WorkerRequest, stream grpc.ServerStreamingServer[pb.Job]) error {
	claimed, err := h.svc.Dequeue(stream.Context(), req.GetWorkerId())
	if err != nil {
		if errors.Is(err, queue.ErrJobNotFound) {
			return status.Error(codes.NotFound, "no jobs available")
		}
		return status.Errorf(codes.Internal, "failed to dequeue job: %v", err)
	}
	pbJob := &pb.Job{
		Id:             claimed.ID,
		Type:           claimed.Type,
		Payload:        claimed.Payload,
		Status:         pb.JobStatus_JOB_STATUS_IN_PROGRESS,
		RetryCount:     claimed.RetryCount,
		MaxRetries:     claimed.MaxRetries,
		IdempotencyKey: claimed.IdempotencyKey.String,
	}

	if err := stream.Send(pbJob); err != nil {
		return status.Errorf(codes.Internal, "failed to send job: %v", err)
	}

	return nil
}

func (h *WorkerHandler) ReportResult(ctx context.Context, req *pb.JobResult) (*pb.Ack, error) {
	if req.GetJobId() == "" {
		return nil, status.Error(codes.InvalidArgument, "job ID is required")
	}

	var err error
	if req.GetSuccess() {
		err = h.svc.Complete(ctx, req.GetJobId())
		h.metrics.JobsCompleted.Inc()
	} else {
		err = h.svc.Fail(ctx, req.GetJobId())
		h.metrics.JobsFailed.Inc()
	}

	if err != nil {
		if errors.Is(err, queue.ErrJobNotFound) {
			return nil, status.Error(codes.NotFound, "job not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to report job result: %v", err)
	}

	return &pb.Ack{Success: true}, nil
}
