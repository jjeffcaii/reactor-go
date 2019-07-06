package flux

import (
	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type Flux interface {
	rs.Publisher
	Filter(rs.Predicate) Flux
	Map(rs.Transformer) Flux
	SubscribeOn(scheduler.Scheduler) Flux
}
