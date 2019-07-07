package mono

import (
	"context"
	"time"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type flatMapper = func(interface{}) Mono

type Mono interface {
	rs.Publisher
	Filter(rs.Predicate) Mono
	Map(rs.Transformer) Mono
	FlatMap(flatMapper) Mono
	SubscribeOn(scheduler.Scheduler) Mono
	Block(context.Context) (interface{}, error)
	DoOnNext(rs.FnOnNext) Mono
	DoOnComplete(rs.FnOnComplete) Mono
	DoOnError(rs.FnOnError) Mono
	DoOnCancel(rs.FnOnCancel) Mono
	DoFinally(rs.FnOnFinally) Mono
	SwitchIfEmpty(alternative Mono) Mono
	DelayElement(delay time.Duration) Mono
}
