package workerpool

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/worker"
	"context"
	"sync"
	"sync/atomic"
)

type Pool struct {
	size    int
	jobs    chan func()
	handler worker.JobHandler
	queue   pb.WorkerServiceClient
	active  atomic.Int64
	sync.WaitGroup
	context.Context
}

func NewPool() {

}
