package queue

import (
	"context"
	"time"
)

type Visibility struct {
	timeout time.Duration
}

func (s *Service) ExtendVisibility(ctx context.Context, jobID string) error {
	computeDeadline := time.Now().Add(s.visibilityCfg.timeout)

	s.repo.ExtendVisibility(ctx, jobID, computeDeadline)
	return
}
