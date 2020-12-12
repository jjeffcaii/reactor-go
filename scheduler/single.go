package scheduler

import (
	"errors"
	"sync"

	"github.com/jjeffcaii/reactor-go/internal/buffer"
)

var errSchedulerClosed = errors.New("scheduler has been closed")

const _singleName = "single"

var (
	_single     Scheduler
	_singleInit sync.Once
)

type singleScheduler struct {
	jobs    *buffer.Unbounded
	started sync.Once
}

func (p *singleScheduler) Name() string {
	return _singleName
}

func (p *singleScheduler) Close() (err error) {
	p.jobs.Dispose()
	return
}

func (p *singleScheduler) start() {
	for {
		p.jobs.Load()
		if next, ok := <-p.jobs.Get(); ok {
			next.(Task)()
		}
	}
}

func (p *singleScheduler) Do(j Task) error {
	if !p.jobs.Put(j) {
		return errSchedulerClosed
	}
	return nil
}

func (p *singleScheduler) Worker() Worker {
	p.started.Do(func() {
		go p.start()
	})
	return p
}

// NewSingle creates a new single scheduler with customized cap.
func NewSingle() Scheduler {
	return &singleScheduler{
		jobs: buffer.NewUnbounded(),
	}
}

// Single returns a scheduler which schedule tasks on a single goroutine.
// It will execute tasks one by one.
func Single() Scheduler {
	_singleInit.Do(func() {
		_single = NewSingle()
	})
	return _single
}
