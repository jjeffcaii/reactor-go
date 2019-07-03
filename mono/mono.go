package mono

import (
	"context"

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
}
