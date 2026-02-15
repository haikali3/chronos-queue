package job

import (
	"encoding/json"
	"errors"
)

var (
	ErrEmptyType      = errors.New("job type cannot be empty")
	ErrInvalidRetries = errors.New("max retried must be equal or greater than 0")
	ErrInvalidPayload = errors.New("payload must be valid JSON")
)

func ValidateJob(job *Job) error {
	if job.Type == "" {
		return ErrEmptyType
	}
	if job.MaxRetries < 0 {
		return ErrInvalidRetries
	}
	if len(job.Payload) > 0 && !json.Valid(job.Payload) {
		return ErrInvalidPayload
	}
	return nil
}
