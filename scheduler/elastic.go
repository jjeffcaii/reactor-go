package scheduler

import (
	"math"
	"sync"

	"github.com/panjf2000/ants/v2"
)

var (
	_elastic     Scheduler
	_elasticInit sync.Once
)

type elasticScheduler struct {
	pool *ants.Pool
}

func (e *elasticScheduler) Close() (err error) {
	e.pool.Release()
	return
}

func (e *elasticScheduler) Do(job Task) {
	if err := e.pool.Submit(job); err != nil {
		panic(err)
	}
}

func (e *elasticScheduler) Worker() Worker {
	return e
}

// NewElastic creates a new elastic scheduler.
func NewElastic(size int) Scheduler {
	pool, _ := ants.NewPool(size)
	return &elasticScheduler{
		pool: pool,
	}
}

// Elastic is a dynamic alloc scheduler.
// It's based on ants goroutine pool.
func Elastic() Scheduler {
	_elasticInit.Do(func() {
		_elastic = NewElastic(math.MaxInt32)
	})
	return _elastic
}
