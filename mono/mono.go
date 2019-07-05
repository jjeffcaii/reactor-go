package mono

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type flatMapper = func(interface{}) Mono

type raw interface {
	SubscribeWith(context.Context, rs.Subscriber)
}

type Mono interface {
	rs.Publisher
	Filter(rs.Predicate) Mono
	Map(rs.Transformer) Mono
	FlatMap(flatMapper) Mono
	SubscribeOn(scheduler.Scheduler) Mono
	Block(context.Context) (interface{}, error)
	DoOnNext(rs.FnOnNext) Mono
	DoOnError(rs.FnOnError) Mono
	DoOnComplete(rs.FnOnComplete) Mono
	DoFinally(rs.FnOnFinally) Mono
	SwitchIfEmpty(alternative Mono) Mono
}
