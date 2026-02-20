package workerpool

func (p *Pool) Utilization() float64 {
	// Returns len(p.jobs) / cap(p.jobs) as a float from 0.0 to 1.0
	// 0.0 = buffer empty (workers keeping up)
	// 1.0 = buffer full (backpressure active)
	return float64(len(p.jobs)) / float64(cap(p.jobs))
}

func (p *Pool) ActiveWorkers() int64 {
	return p.active.Load()
}

func (p *Pool) IdleWorkers() int64 {
	return p.idle.Load()
}

func (p *Pool) BufferLen() int {
	return len(p.jobs)
}

func (p *Pool) BufferCap() int {
	return cap(p.jobs)
}

func (p *Pool) IsSaturated() bool {
	return p.Utilization() >= 0.9
}
