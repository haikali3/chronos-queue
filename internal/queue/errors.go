package queue

import "errors"

var (
	ErrJobNotFound             = errors.New("job not found")
	ErrInvalidTransition       = errors.New("invalid job state transition")
	ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")
)
