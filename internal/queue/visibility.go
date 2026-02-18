package queue

import (
	"context"
	"time"
)

type VisibilityConfig struct {
	timeout time.Duration
}

func (s *Service) ExtendVisibility(ctx context.Context, jobID string) error {
	visibleAfter := time.Now().Add(s.visibilityCfg.timeout)
	return s.repo.ExtendVisibility(ctx, jobID, visibleAfter)
}
