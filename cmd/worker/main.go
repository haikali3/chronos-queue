package main

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/config"
	"chronos-queue/internal/logger"
	"chronos-queue/internal/worker"
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger.Init()
	log := logger.Get()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	queueAddr := fmt.Sprintf("localhost:%d", cfg.QueueGRPCPort)
	conn, err := grpc.NewClient(queueAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect to queue service", zap.Error(err))
	}
	defer func() { _ = conn.Close() }()

	queueClient := pb.NewWorkerServiceClient(conn)
	handler := &worker.SimulatedHandler{}
	workerID := uuid.New().String()
	w := worker.New(workerID, queueClient, handler, cfg.WorkerPollInterval)

	log.Info("starting worker", zap.String("worker_id", workerID), zap.String("queue_addr", queueAddr))
	w.Run(context.Background())
}
