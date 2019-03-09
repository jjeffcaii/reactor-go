package rs

import "context"

var (

	elastic   = &elasticScheduler{}
	immediate = &immediateScheduler{}
)

type elasticScheduler struct {
}

func (*elasticScheduler) Do(ctx context.Context, fn func(ctx context.Context)) {
	go fn(ctx)
}

type immediateScheduler struct {
}

func (*immediateScheduler) Do(ctx context.Context, fn func(ctx context.Context)) {
	fn(ctx)
}

func Elastic() Scheduler {
	return elastic
}

func Immediate() Scheduler {
	return immediate
}
