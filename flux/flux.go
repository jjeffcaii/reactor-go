package flux

import (
	"github.com/jjeffcaii/reactor-go"
)

type Sink interface {
	Next(v interface{}) error
	Error(e error)
	Complete()
}

type Flux interface {
	rs.Publisher
	Filter(fn rs.Predicate) Flux
	Map(fn rs.FnTransform) Flux
	SubscribeOn(s rs.Scheduler) Flux
	PublishOn(s rs.Scheduler) Flux
}
