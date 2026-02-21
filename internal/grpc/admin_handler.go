package grpc

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/observability"
	"chronos-queue/internal/queue"
	"context"
)

type AdminHandler struct {
	svc     *queue.Service
	metrics *observability.Metrics
	pb.UnimplementedAdminServiceServer
}

func NewAdminHandler(svc *queue.Service, metrics *observability.Metrics) *AdminHandler {
	return &AdminHandler{
		svc:     svc,
		metrics: metrics,
	}
}

func (h *AdminHandler) ListDeadLetterJobs(ctx context.Context, req *pb.ListDLQRequest) (*pb.ListDLQResponse, error) {
	jobs, total, err := h.svc.ListDeadLetterJobs(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	pbJobs := make([]*pb.Job, len(jobs))
	for i, job := range jobs {
		pbJobs[i] = &pb.Job{
			Id:             job.ID,
			Type:           job.Type,
			Payload:        job.Payload,
			Status:         pb.JobStatus(pb.JobStatus_value[job.Status]),
			RetryCount:     job.RetryCount,
			MaxRetries:     job.MaxRetries,
			IdempotencyKey: job.IdempotencyKey.String,
			DlqReason:      job.DlqReason.String,
		}
	}

	return &pb.ListDLQResponse{
		Jobs:  pbJobs,
		Total: int32(total),
	}, nil
}

func (h *AdminHandler) GetJobDetails(ctx context.Context, req *pb.GetJobRequest) (*pb.GetJobResponse,
	error) {
	job, err := h.svc.GetJobDetails(ctx, req.JobId)
	if err != nil {
		return nil, err
	}
	return &pb.GetJobResponse{
		Job: &pb.Job{
			Id:             job.ID,
			Type:           job.Type,
			Payload:        job.Payload,
			Status:         pb.JobStatus(pb.JobStatus_value[job.Status]),
			RetryCount:     job.RetryCount,
			MaxRetries:     job.MaxRetries,
			IdempotencyKey: job.IdempotencyKey.String,
			DlqReason:      job.DlqReason.String,
		},
	}, nil
}

func (h *AdminHandler) RetryJob(ctx context.Context, req *pb.RetryJobRequest) (*pb.RetryJobResponse, error) {
	job, err := h.svc.RetryJob(ctx, req.JobId)
	if err != nil {
		return nil, err
	}
	return &pb.RetryJobResponse{
		Job: &pb.Job{
			Id:             job.ID,
			Type:           job.Type,
			Payload:        job.Payload,
			Status:         pb.JobStatus(pb.JobStatus_value[job.Status]),
			RetryCount:     job.RetryCount,
			MaxRetries:     job.MaxRetries,
			IdempotencyKey: job.IdempotencyKey.String,
			DlqReason:      job.DlqReason.String,
		},
	}, nil
}
