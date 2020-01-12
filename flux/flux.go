package flux

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type OverflowStrategy int8

const (
	OverflowBuffer OverflowStrategy = iota
	OverflowIgnore
	OverflowError
	OverflowDrop
	OverflowLatest
)

type FnSwitchOnFirst = func(s Signal, f Flux) Flux

type Flux interface {
	rs.Publisher
	Filter(rs.Predicate) Flux
	Map(rs.Transformer) Flux
	Take(n int) Flux
	DoOnDiscard(rs.FnOnDiscard) Flux
	DoOnNext(rs.FnOnNext) Flux
	DoOnComplete(rs.FnOnComplete) Flux
	DoOnError(rs.FnOnError) Flux
	DoOnCancel(rs.FnOnCancel) Flux
	DoOnRequest(rs.FnOnRequest) Flux
	DoOnSubscribe(rs.FnOnSubscribe) Flux
	DoFinally(rs.FnOnFinally) Flux
	SwitchOnFirst(FnSwitchOnFirst) Flux
	DelayElement(delay time.Duration) Flux
	SubscribeOn(scheduler.Scheduler) Flux
	BlockFirst(context.Context) (interface{}, error)
	BlockLast(context.Context) (interface{}, error)
	ToChan(ctx context.Context, cap int) (c <-chan interface{}, e <-chan error)
}

type Sink interface {
	Complete()
	Error(error)
	Next(interface{})
}

type Processor interface {
	Flux
	Sink
}
