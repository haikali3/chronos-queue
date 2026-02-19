package workerpool

import (
	"chronos-queue/gen/pb"
	"chronos-queue/internal/logger"
	"chronos-queue/internal/worker"
	"context"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Pool = the workers. It's a group of N goroutines sitting idle, waiting for jobs to show up on a channel. When a job arrives, one
// goroutine picks it up, runs the handler, reports the result back via gRPC, then goes back to waiting. It doesn't know or care where
// jobs come from.

// Worker goroutine logic:
// 1. Increment idle
// 2. select on ctx.Done() (exit) or <-jobs (got work)
// 3. On job received: decrement idle, increment active
// 4. Call handler.Handle(ctx, job.Proto)
// 5. Call queue.ReportResult() with success/failure
// 6. Decrement active, loop back

type Job struct {
	Proto    *pb.Job
	WorkerID string
	Ctx      context.Context
	Span     trace.Span
}

type Pool struct {
	size    int
	jobs    chan Job
	handler worker.JobHandler
	queue   pb.WorkerServiceClient

	active atomic.Int64
	idle   atomic.Int64

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewPool(size int, bufferSize int, handler worker.JobHandler, queue pb.WorkerServiceClient) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		size:    size,
		jobs:    make(chan Job, bufferSize),
		handler: handler,
		queue:   queue,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (p *Pool) Start() {
	log := logger.Get()
	log.Info("starting worker pool", zap.Int("size", p.size))

	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *Pool) Submit(job Job) bool {
	select {

	case p.jobs <- job:
		return true
	case <-p.ctx.Done():
		// Pool is stopping, reject new jobs
		logger.Get().Warn("worker pool is stopping, rejecting new job")
		return false
	}
}

func (p *Pool) Jobs() chan<- Job {
	return p.jobs
}

func (p *Pool) Stop() {
	log := logger.Get()
	log.Info("stopping worker pool, waiting for in-progress jobs to finish")
	p.cancel()
	p.wg.Wait()
	log.Info("worker pool stopped")
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	log := logger.Get()
	log.Info("pool worker started", zap.Int("worker_num", id))

	for {
		p.idle.Add(1)
		select {
		case <-p.ctx.Done():
			p.idle.Add(-1)
			log.Info("worker stopping", zap.Int("worker_num", id))
			return
		case job, ok := <-p.jobs:
			p.idle.Add(-1)
			if !ok {
				log.Info("job channel closed, worker stopping", zap.Int("worker_num", id))
				return
			}
			p.active.Add(1)
			log.Info("worker executing job", zap.Int("worker_num", id))
			p.process(job)
			p.active.Add(-1)
		}
	}
}

func (p *Pool) process(job Job) {
	defer job.Span.End()
	log := logger.Get()
	jobID := job.Proto.GetId()

	err := p.handler.Handle(job.Ctx, job.Proto)

	success := true
	if err != nil {
		log.Error("job failed", zap.String("job_id", jobID), zap.Error(err))
		success = false
	}

	_, reportErr := p.queue.ReportResult(job.Ctx, &pb.JobResult{
		JobId:    jobID,
		WorkerId: job.WorkerID,
		Success:  success,
	})
	if reportErr != nil {
		log.Error("failed to report job result", zap.String("job_id", jobID), zap.Error(reportErr))
	}
}
