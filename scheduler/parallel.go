package scheduler

import (
  "runtime"
  "sync"
)

var parallel Scheduler

func init() {
  parallel = NewParallel(runtime.NumCPU() * 2)
}

type parallelScheduler struct {
  jobs chan Job
  n    int
  once sync.Once
}

func (p *parallelScheduler) Close() error {
  close(p.jobs)
  return nil
}

func (p *parallelScheduler) Do(j Job) {
  p.jobs <- j
}

func (p *parallelScheduler) start() {
  for i := 0; i < p.n; i++ {
    go func() {
      for j := range p.jobs {
        j()
      }
    }()
  }
}

func (p *parallelScheduler) Worker() Worker {
  p.once.Do(func() {
    p.start()
  })
  return p
}

func NewParallel(n int) Scheduler {
  return &parallelScheduler{
    jobs: make(chan Job),
    n:    n,
  }
}

func Parallel() Scheduler {
  return parallel
}
