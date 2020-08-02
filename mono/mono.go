package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type Any = reactor.Any
type flatMapper = func(reactor.Any) Mono

type Mono interface {
	reactor.Publisher
	Filter(reactor.Predicate) Mono
	Map(reactor.Transformer) Mono
	FlatMap(flatMapper) Mono
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
	DelayElement(delay time.Duration) Mono
}

type Processor interface {
	Mono
	Sink
}
