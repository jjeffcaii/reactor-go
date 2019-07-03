package flux

import (
  "context"

  "github.com/jjeffcaii/reactor-go"
  "github.com/jjeffcaii/reactor-go/scheduler"
)

type fluxCreate struct {
  source       func(context.Context, Sink)
  backpressure OverflowStrategy
}

func (m fluxCreate) SubscribeOn(sc scheduler.Scheduler) Flux {
  return newFluxSubscribeOn(m, sc)
}

func (m fluxCreate) Subscribe(ctx context.Context, s rs.Subscriber) {
  var sink Sink
  switch m.backpressure {
  case OverflowBuffer:
    sink = newBufferedSink(s, 16)
  default:
    panic("not implement")
  }
  s.OnSubscribe(sink)
  m.source(ctx, sink)
}

func (m fluxCreate) Filter(predicate rs.Predicate) Flux {
  return newFluxFilter(m, predicate)
}

func (m fluxCreate) Map(tf rs.Transformer) Flux {
  return newFluxMap(m, tf)
}

func Create(c func(ctx context.Context, sink Sink)) Flux {
  return &fluxCreate{
    source:       c,
    backpressure: OverflowBuffer,
  }
}
