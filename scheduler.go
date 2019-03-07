package rs

import "context"

type elasticScheduler struct {
}

func (*elasticScheduler) Do(ctx context.Context, fn func(ctx context.Context)) {
  go fn(ctx)
}

type emmSche struct {
}

func (*emmSche) Do(ctx context.Context, fn func(ctx context.Context)) {
  fn(ctx)
}

var elastic = &elasticScheduler{}

func Elastic() Scheduler {
  return elastic
}

var emm = &emmSche{}

func Immet() Scheduler {
  return emm
}
