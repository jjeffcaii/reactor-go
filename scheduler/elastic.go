package scheduler

import (
	"math"
	"sync"

	"github.com/panjf2000/ants/v2"
)

const _elasticName = "elastic"
const _elasticBoundedSize = 10000

var (
	_elastic            Scheduler
	_elasticInit        sync.Once
	_elasticBounded     Scheduler
	_elasticBoundedInit sync.Once
)

type elasticScheduler struct {
	pool *ants.Pool
}

func (e *elasticScheduler) Name() string {
	return _elasticName
}

func (e *elasticScheduler) Close() (err error) {
	e.pool.Release()
	return
}

func (e *elasticScheduler) Do(job Task) error {
	return e.pool.Submit(job)
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

// ElasticBounded is a dynamic scheduler with bounded size.
// It's based on ants goroutine pool.
func ElasticBounded() Scheduler {
	_elasticBoundedInit.Do(func() {
		_elasticBounded = NewElastic(_elasticBoundedSize)
	})
	return _elasticBounded
}
