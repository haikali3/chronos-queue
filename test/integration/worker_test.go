package integration

import (
	"chronos-queue/internal/queue"
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestMultiWorker(t *testing.T) {
	svc, pool := setupService(t)
	defer pool.Close()

	ctx := context.Background()

	for i := range 100 {
		_, err := svc.Enqueue(
			ctx,
			"multi-worker-test-job",
			fmt.Appendf(nil, "payload-%d", i),
			3,
			"",
		)
		if err != nil {
			t.Fatalf("Failed to enqueue job %d: %v", i, err)
		}
	}

	var wg sync.WaitGroup
	var claimed sync.Map
	for i := range 3 { // spawn 3 workers
		wg.Add(1)
		go func(workerID string) {
			defer wg.Done()
			for {
				job, err := svc.Dequeue(ctx, workerID)
				if err == queue.ErrJobNotFound {
					break
				}
				if _, loaded := claimed.LoadOrStore(job.ID, workerID); loaded {
					t.Errorf("Worker %s claimed job %s that was already claimed", workerID, job.ID)
				}
				err = svc.Complete(ctx, job.ID)
				if err != nil {
					t.Errorf("Worker %s failed to dequeue job: %v", workerID, err)
					return
				}
			}
		}(fmt.Sprintf("worker-%d", i))
	}
	wg.Wait()

	// count claimed jobs
	var totalClaimed int
	claimed.Range(func(key, value any) bool {
		totalClaimed++
		return true
	})
	if totalClaimed != 100 {
		t.Errorf("Expected 100 claimed jobs, got %d", totalClaimed)
	}

	// verify db
	var completedCount int
	err := pool.QueryRow(ctx, "SELECT count(*) FROM jobs WHERE status = 'COMPLETED'").Scan(&completedCount)
	if err != nil {
		t.Fatalf("Failed to query completed jobs: %v", err)
	}
	if completedCount != 100 {
		t.Errorf("Expected 100 completed jobs, got %d", completedCount)
	}
}
