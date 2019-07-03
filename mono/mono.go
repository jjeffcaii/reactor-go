package mono

import (
	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type Mono interface {
	rs.Publisher
	Filter(rs.Predicate) Mono
	Map(rs.Transformer) Mono
	SubscribeOn(scheduler.Scheduler) Mono
}
