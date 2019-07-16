package flux

import (
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type Flux interface {
	rs.Publisher
	Filter(rs.Predicate) Flux
	Map(rs.Transformer) Flux
	DoOnDiscard(rs.FnOnDiscard) Flux
	DoOnNext(rs.FnOnNext) Flux
	DoOnComplete(rs.FnOnComplete) Flux
	DoOnRequest(rs.FnOnRequest) Flux
	DoFinally(rs.FnOnFinally) Flux
	SubscribeOn(scheduler.Scheduler) Flux
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
