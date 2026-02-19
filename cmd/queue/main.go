package main

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/config"
	"chronos-queue/internal/logger"
	"chronos-queue/internal/observability"
	"chronos-queue/internal/queue"
	"chronos-queue/internal/storage/postgres"
	"context"
	"fmt"
	"net"
	"net/http"

	grpchandler "chronos-queue/internal/grpc"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger.Init()
	log := logger.Get()
	defer logger.Sync()

	tp, err := observability.InitTracer(context.Background(), "chronos-queue")
	if err != nil {
		log.Fatal("failed to initialize tracer", zap.Error(err))
	}
	defer observability.ShutdownTracer(tp)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to db", zap.Error(err))
	}
	defer pool.Close()

	repo := postgres.NewJobRepository(pool)
	svc := queue.New(repo, log)

	metrics := observability.NewMetrics()
	metricsAddr := fmt.Sprintf(":%d", cfg.MetricsPort)

	go func() {
		http.Handle("/metrics", metrics.Handler())
		log.Info("Starting metrics server", zap.String("address", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			log.Error("failed to start metrics server", zap.Error(err))
		}
	}()

	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterProducerServiceServer(grpcServer, grpchandler.NewProducerHandler(svc, metrics))
	pb.RegisterWorkerServiceServer(grpcServer, grpchandler.NewWorkerHandler(svc, metrics))

	addr := fmt.Sprintf(":%d", cfg.QueueGRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to listen on port", zap.Int("port", cfg.QueueGRPCPort), zap.Error(err))
	}

	log.Info("Queue service gRPC server starting", zap.String("address", addr))
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("failed to start gRPC server", zap.Error(err))
	}

}
