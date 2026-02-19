package main

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/config"
	"chronos-queue/internal/logger"
	"chronos-queue/internal/observability"
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type producerGateway struct {
	pb.UnimplementedProducerServiceServer
	queue pb.ProducerServiceClient
}

func (g *producerGateway) SubmitJob(ctx context.Context, req *pb.JobRequest) (*pb.JobResponse, error) {
	log := logger.Get()

	if req.GetType() == "" {
		return nil, status.Error(codes.InvalidArgument, "job type is required")
	}

	log.Info("forwarding job", zap.String("type", req.GetType()))

	resp, err := g.queue.SubmitJob(ctx, req)
	if err != nil {
		log.Error("queue service error", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

func main() {
	logger.Init()
	log := logger.Get()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	tp, err := observability.InitTracer(context.Background(), "chronos-producer")
	if err != nil {
		log.Fatal("failed to initialize tracer", zap.Error(err))
	}
	defer observability.ShutdownTracer(tp)

	queueAddr := fmt.Sprintf("localhost:%d", cfg.QueueGRPCPort)
	conn, err := grpc.NewClient(queueAddr,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("failed to connect to queue service", zap.Error(err))
	}
	defer func() { _ = conn.Close() }()

	queueClient := pb.NewProducerServiceClient(conn)
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterProducerServiceServer(grpcServer, &producerGateway{
		queue: queueClient,
	})

	addr := fmt.Sprintf(":%d", cfg.ProducerGRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to listen", zap.String("addr", addr), zap.Error(err))
	}

	log.Info("starting producer gateway", zap.String("addr", addr), zap.String("queue", queueAddr))
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("failed to start gRPC server", zap.Error(err))
	}
}
