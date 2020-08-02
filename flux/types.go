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

type Any = reactor.Any
type FnSwitchOnFirst = func(s Signal, f Flux) Flux

type Flux interface {
	reactor.Publisher
	Filter(reactor.Predicate) Flux
	Map(reactor.Transformer) Flux
	Take(n int) Flux
	DoOnDiscard(reactor.FnOnDiscard) Flux
	DoOnNext(reactor.FnOnNext) Flux
	DoOnComplete(reactor.FnOnComplete) Flux
	DoOnError(reactor.FnOnError) Flux
	DoOnCancel(reactor.FnOnCancel) Flux
	DoOnRequest(reactor.FnOnRequest) Flux
	DoOnSubscribe(reactor.FnOnSubscribe) Flux
	DoFinally(reactor.FnOnFinally) Flux
	SwitchOnFirst(FnSwitchOnFirst) Flux
	DelayElement(delay time.Duration) Flux
	SubscribeOn(scheduler.Scheduler) Flux
	SubscribeWithChan(ctx context.Context, valueChan interface{}, errChan chan<- error)
	BlockFirst(context.Context) (Any, error)
	BlockLast(context.Context) (Any, error)
	BlockToSlice(ctx context.Context, slicePtr interface{}) error
}

type Sink interface {
	Complete()
	Error(error)
	Next(Any)
}

type Processor interface {
	Flux
	Sink
}
