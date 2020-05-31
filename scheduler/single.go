package scheduler

import (
	"sync"
	"sync/atomic"
)

var (
	_single     Scheduler
	_singleInit sync.Once
)

type singleScheduler struct {
	jobs    chan Job
	started int64
}

func (p *singleScheduler) Close() (err error) {
	close(p.jobs)
	return
}

func (p *singleScheduler) start() {
	for j := range p.jobs {
		j()
	}
}

func (p *singleScheduler) Do(j Job) {
	p.jobs <- j
}

func (p *singleScheduler) Worker() Worker {
	if atomic.AddInt64(&(p.started), 1) == 1 {
		go p.start()
	}
	return p
}

func NewSingle(cap int) Scheduler {
	return &singleScheduler{
		jobs: make(chan Job, cap),
	}
}

func Single() Scheduler {
	_singleInit.Do(func() {
		_single = NewSingle(1000)
	})
	return _single
}
