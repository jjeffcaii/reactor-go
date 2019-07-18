package scheduler

import (
	"math"

	"github.com/panjf2000/ants"
)

var elastic Scheduler

func init() {
	pool, err := ants.NewPool(math.MaxInt32)
	if err != nil {
		panic(err)
	}
	elastic = &elasticScheduler{
		pool: pool,
	}
}

type elasticScheduler struct {
	pool *ants.Pool
}

func (p *elasticScheduler) Do(job Job) {
	if err := p.pool.Submit(job); err != nil {
		panic(err)
	}
}

func (p *elasticScheduler) Worker() Worker {
	return p
}

func Elastic() Scheduler {
	return elastic
}
