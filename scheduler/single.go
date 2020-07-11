package scheduler

import (
	"sync"
	"sync/atomic"
)

// DefaultSingleCap is default task queue size.
const DefaultSingleCap = 1000

var (
	_single     Scheduler
	_singleInit sync.Once
)

type singleScheduler struct {
	jobs    chan Task
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

func (p *singleScheduler) Do(j Task) {
	p.jobs <- j
}

func (p *singleScheduler) Worker() Worker {
	if atomic.AddInt64(&(p.started), 1) == 1 {
		go p.start()
	}
	return p
}

// NewSingle creates a new single scheduler with customized cap.
func NewSingle(cap int) Scheduler {
	return &singleScheduler{
		jobs: make(chan Task, cap),
	}
}

// Single returns a scheduler which schedule tasks on a single goroutine.
// It will execute tasks one by one.
func Single() Scheduler {
	_singleInit.Do(func() {
		_single = NewSingle(DefaultSingleCap)
	})
	return _single
}
