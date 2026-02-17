package main

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/config"
	"chronos-queue/internal/logger"
	"chronos-queue/internal/worker"
	"chronos-queue/internal/workerpool"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Connect to queue service via grpc
	queueAddr := fmt.Sprintf("localhost:%d", cfg.QueueGRPCPort)
	conn, err := grpc.NewClient(queueAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect to queue service", zap.Error(err))
	}
	defer func() { _ = conn.Close() }()

	queueClient := pb.NewWorkerServiceClient(conn)
	handler := &worker.SimulatedHandler{}
	workerID := uuid.New().String()

	pool := workerpool.NewPool(cfg.WorkerPoolSize, cfg.WorkerBufferSize, handler, queueClient)

	log.Info("starting worker", zap.String("worker_id", workerID), zap.String("queue_addr", queueAddr))
	pool.Start()

	dispatcher := workerpool.NewDispatcher(pool, queueClient, workerID, cfg.WorkerPollInterval)
	go dispatcher.Start()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Info("shutdown signal received, initiating graceful shutdown")
	workerpool.GracefulShutdown(dispatcher, pool, 30*time.Second)
}
