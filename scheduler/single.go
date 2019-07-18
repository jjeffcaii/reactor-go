package scheduler

import "sync/atomic"

var (
  single            Scheduler
  defaultSingleJobs = 1000
)

func init() {
  single = NewSingle(defaultSingleJobs)
}

func NewSingle(cap int) Scheduler {
  return &singleScheduler{
    jobs: make(chan Job, cap),
  }
}

type singleScheduler struct {
  jobs    chan Job
  started int64
}

func (p *singleScheduler) Close() (err error) {
  close(p.jobs)
  return
}

func (p *singleScheduler) start() {
L:
  for {
    select {
    case j, ok := <-p.jobs:
      if !ok {
        break L
      }
      j()
    }
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

func Single() Scheduler {
  return single
}
