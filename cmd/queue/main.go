package main

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/config"
	"chronos-queue/internal/logger"
	"chronos-queue/internal/queue"
	"chronos-queue/internal/storage/postgres"
	"context"
	"fmt"
	"net"

	grpchandler "chronos-queue/internal/grpc"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger.Init()
	log := logger.Get()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to create database pool", zap.Error(err))
	}
	defer pool.Close()

	repo := postgres.NewJobRepository(pool)
	svc := queue.New(repo, log)

	srv := grpc.NewServer()
	pb.RegisterProducerServiceServer(srv, grpchandler.NewProducerHandler(svc))
	pb.RegisterWorkerServiceServer(srv, grpchandler.NewWorkerHandler(svc))

	addr := fmt.Sprintf(":%d", cfg.QueueGRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to listen", zap.String("addr", addr), zap.Error(err))
	}

	log.Info("starting gRPC server", zap.String("addr", addr))
	if err := srv.Serve(listener); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}
