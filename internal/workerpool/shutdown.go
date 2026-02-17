package workerpool

func (p *WorkerPool) Shutdown() {
	// 1. stop polling for new jobs
	p.dispatcher.Stop()

	// 2. drain remaining buffered jobs
	p.close(pool.jobs)
	// 3. wait on pool.wg with timeout:
	//    - spawn goroutine that calls pool.wg.Wait(), signal done channel
	//    - select on done channel or time.After(timeout)
	//    - if timeout: log warning, cancel pool context (force kill)
	// 4. pool.Stop()
}
