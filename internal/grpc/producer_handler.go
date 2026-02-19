package grpc

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/observability"
	"chronos-queue/internal/queue"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProducerHandler struct {
	pb.UnimplementedProducerServiceServer
	svc     *queue.Service         // svc = service
	metrics *observability.Metrics // prometheus,opentelemetry metrics and jaegar metrics
}

func NewProducerHandler(svc *queue.Service, metrics *observability.Metrics) *ProducerHandler {
	return &ProducerHandler{svc: svc, metrics: metrics}
}

func (h *ProducerHandler) SubmitJob(ctx context.Context, req *pb.JobRequest) (*pb.JobResponse, error) {
	if req.GetType() == "" {
		return nil, status.Error(codes.InvalidArgument, "job type is required")
	}

	created, err := h.svc.Enqueue(ctx, req.GetType(), req.GetPayload(), req.GetMaxRetries(), req.GetIdempotencyKey())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to enqueue job: %v", err)
	}
	h.metrics.JobsSubmitted.Inc()

	return &pb.JobResponse{
		Id:     created.ID,
		Status: pb.JobStatus_JOB_STATUS_PENDING,
	}, nil
}
