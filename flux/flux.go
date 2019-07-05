package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type raw interface {
	SubscribeWith(context.Context, rs.Subscriber)
}

type Flux interface {
	rs.Publisher
	Filter(rs.Predicate) Flux
	Map(rs.Transformer) Flux
	SubscribeOn(scheduler.Scheduler) Flux
}
