package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	QueueGRPCPort      int
	ProducerGRPCPort   int
	WorkerPollInterval time.Duration
	WorkerPoolSize     int
	WorkerBufferSize   int
	RedisURL           string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")

	queuePort, err := getEnvInt("QUEUE_GRPC_PORT", 50051)
	if err != nil {
		return cfg, fmt.Errorf("invalid QUEUE_GRPC_PORT: %v", err)
	}
	cfg.QueueGRPCPort = queuePort

	producerPort, err := getEnvInt("PRODUCER_GRPC_PORT", 50052)
	if err != nil {
		return cfg, fmt.Errorf("invalid PRODUCER_GRPC_PORT: %v", err)
	}
	cfg.ProducerGRPCPort = producerPort

	pollMs, err := getEnvInt("WORKER_POLL_INTERVAL_MS", 1000)
	if err != nil {
		return cfg, fmt.Errorf("invalid WORKER_POLL_INTERVAL_MS: %w", err)
	}
	cfg.WorkerPollInterval = time.Duration(pollMs) * time.Millisecond

	poolSize, err := getEnvInt("WORKER_POOL_SIZE", 10)
	if err != nil {
		return cfg, fmt.Errorf("invalid WORKER_POOL_SIZE: %w", err)
	}
	cfg.WorkerPoolSize = poolSize

	bufferSize, err := getEnvInt("WORKER_BUFFER_SIZE", 100)
	if err != nil {
		return cfg, fmt.Errorf("invalid WORKER_BUFFER_SIZE: %w", err)
	}
	cfg.WorkerBufferSize = bufferSize
	cfg.RedisURL = os.Getenv("REDIS_URL")

	if err := validate(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func getEnvInt(key string, defaultValue int) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(valueStr)
}
