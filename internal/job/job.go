package job

import "time"

type Job struct {
	ID             string
	Type           string
	Payload        []byte
	Status         JobStatus
	RetryCount     int32
	MaxRetries     int32
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
