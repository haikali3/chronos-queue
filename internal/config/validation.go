package config

import "fmt"

func validate(cfg Config) error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.QueueGRPCPort <= 0 || cfg.QueueGRPCPort > 65535 {
		return fmt.Errorf("QUEUE_GRPC_PORT must be between 1 and 65535, got %d", cfg.QueueGRPCPort)
	}
	if cfg.ProducerGRPCPort <= 0 || cfg.ProducerGRPCPort > 65535 {
		return fmt.Errorf("PRODUCER_GRPC_PORT must be between 1 and 65535, got %d", cfg.ProducerGRPCPort)
	}
	if cfg.QueueGRPCPort == cfg.ProducerGRPCPort {
		return fmt.Errorf("QUEUE_GRPC_PORT and PRODUCER_GRPC_PORT cannot be the same, got %d", cfg.QueueGRPCPort)
	}
	if cfg.WorkerPollInterval <= 0 {
		return fmt.Errorf("WORKER_POLL_INTERVAL_MS must be greater than 0, got %d", cfg.WorkerPollInterval.Milliseconds())
	}
	if cfg.WorkerPoolSize <= 0 {
		return fmt.Errorf("WORKER_POOL_SIZE must be greater than 0, got %d", cfg.WorkerPoolSize)
	}
	if cfg.WorkerBufferSize < 0 {
		return fmt.Errorf("WORKER_BUFFER_SIZE cannot be negative, got %d", cfg.WorkerBufferSize)
	}
	return nil
}
