package scheduler

import (
  "context"
  "errors"
  "math"
  "sync"
  "sync/atomic"
)

var myElastic = elasticScheduler{}

type elasticScheduler struct {
}

func (elasticScheduler) Worker() Worker {
  return &singleWorker{
    jobs: make(chan Job),
  }
}

type singleWorker struct {
  accepts int64
  jobs    chan Job
  once    sync.Once
  closed  bool
}

func (s *singleWorker) Close() (err error) {
  s.once.Do(func() {
    atomic.StoreInt64(&(s.accepts), math.MinInt64)
    close(s.jobs)
  })
  return
}

func (s *singleWorker) Do(job Job) {
  cur := atomic.AddInt64(&(s.accepts), 1)
  if cur < 1 {
    panic(errors.New("worker has been stopped"))
  }
  if cur == 1 {
    go s.loop(context.Background())
  }
  s.jobs <- job
}

func (s *singleWorker) loop(ctx context.Context) {
  for {
    select {
    case <-ctx.Done():
      _ = s.Close()
      goto EXIT
    case job, ok := <-s.jobs:
      if !ok {
        return
      }
      job()
    }
  }
EXIT:
  for job := range s.jobs {
    job()
  }
}

func Elastic() Scheduler {
  return myElastic
}
