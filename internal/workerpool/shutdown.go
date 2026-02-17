package workerpool

import (
	"chronos-queue/internal/logger"
	"time"
)

func GracefulShutdown(dispatcher *Dispatcher, pool *Pool, timeout time.Duration) {
	log := logger.Get()

	// 1. stop polling for new jobs
	dispatcher.Stop()
	log.Info("dispatcher stopped, no longer polling for new jobs")

	// 2. drain remaining buffered jobs
	close(pool.jobs)
	log.Info("jobs channel closed, draining in-flight jobs")

	// 3. wait on pool.wg with timeout:
	//    - spawn goroutine that calls pool.wg.Wait(), signal done channel
	//    - select on done channel or time.After(timeout)
	//    - if timeout: log warning, cancel pool context (force kill)
	done := make(chan struct{})
	go func() {
		pool.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("all in-flight jobs completed, shutting down worker pool")
	case <-time.After(timeout):
		log.Warn("graceful shutdown timed out, force stopping worker pool")
		pool.cancel()
	}
	// 4. pool.Stop()
}
