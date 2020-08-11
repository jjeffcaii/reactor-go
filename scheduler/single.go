package scheduler

import (
	"errors"
	"sync"
)

// DefaultSingleCap is default task queue size.
const DefaultSingleCap = 1000

var errSchedulerClosed = errors.New("scheduler has been closed")

const _singleName = "single"

var (
	_single     Scheduler
	_singleInit sync.Once
)

type singleScheduler struct {
	jobs    chan Task
	started sync.Once
}

func (p *singleScheduler) Name() string {
	return _singleName
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

func (p *singleScheduler) Do(j Task) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errSchedulerClosed
		}
	}()
	p.jobs <- j
	return
}

func (p *singleScheduler) Worker() Worker {
	p.started.Do(func() {
		go p.start()
	})
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
