package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

// Alias
type (
	Any        = reactor.Any
	Disposable = reactor.Disposable
)

type FlatMapper = func(reactor.Any) Mono

type Mono interface {
	reactor.Publisher
	Filter(reactor.Predicate) Mono
	Map(reactor.Transformer) Mono
	FlatMap(FlatMapper) Mono
	SubscribeOn(scheduler.Scheduler) Mono
	Block(context.Context) (Any, error)
	DoOnNext(reactor.FnOnNext) Mono
	DoOnComplete(reactor.FnOnComplete) Mono
	DoOnSubscribe(reactor.FnOnSubscribe) Mono
	DoOnError(reactor.FnOnError) Mono
	DoOnCancel(reactor.FnOnCancel) Mono
	DoFinally(reactor.FnOnFinally) Mono
	DoOnDiscard(reactor.FnOnDiscard) Mono
	SwitchIfEmpty(alternative Mono) Mono
	SwitchIfError(alternative func(error) Mono) Mono
	SwitchValueIfError(v Any) Mono
	DelayElement(delay time.Duration) Mono
	Timeout(timeout time.Duration) Mono
	ZipWith(other Mono) Mono
	ZipCombineWith(other Mono, cmb Combinator) Mono
	Raw() reactor.RawPublisher
}

type Processor interface {
	Mono
	Sink
}

type rawProcessor interface {
	reactor.RawPublisher
	Sink
}
