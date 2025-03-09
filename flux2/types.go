package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

const (
	statCancel   = -1
	statError    = -2
	statComplete = 2
)

const (
	flagCancel uint8 = 1 << iota
	flagFast
)

type OverflowStrategy int8

const (
	OverflowBuffer OverflowStrategy = iota
	OverflowIgnore
	OverflowError
	OverflowDrop
	OverflowLatest
)

func Map[A, B any](origin Flux[A], tf reactor.Transformer[A, B]) Flux[B] {
	return wrap[B](newFluxMap[A, B](origin.Raw(), tf))
}

type Flux[T any] interface {
	reactor.Publisher[T]

	Filter(reactor.Predicate[T]) Flux[T]

	Map(reactor.Transformer[T, T]) Flux[T]

	Take(n int) Flux[T]

	//DoOnDiscard(reactor.FnOnDiscard) Flux

	DoOnNext(reactor.FnOnNext[T]) Flux[T]

	DoOnComplete(reactor.FnOnComplete) Flux[T]

	DoOnError(reactor.FnOnError) Flux[T]

	DoOnCancel(reactor.FnOnCancel) Flux[T]

	DoOnRequest(reactor.FnOnRequest) Flux[T]

	DoOnSubscribe(reactor.FnOnSubscribe) Flux[T]

	DoFinally(reactor.FnOnFinally) Flux[T]

	//SwitchOnFirst(FnSwitchOnFirst) Flux

	//DelayElement(delay time.Duration) Flux

	SubscribeOn(scheduler.Scheduler) Flux[T]

	//SubscribeWithChan(ctx context.Context, valueChan interface{}, errChan chan<- error)

	BlockFirst(context.Context) (reactor.Null[T], error)

	BlockLast(context.Context) (reactor.Null[T], error)

	//BlockToSlice(ctx context.Context, slicePtr interface{}) error

	Raw() reactor.RawPublisher[T]
}

type Sink[T any] interface {
	Complete(context.Context)
	Error(context.Context, error)
	Next(context.Context, T)
}

type Processor[T any] interface {
	Flux[T]
	Sink[T]
}
