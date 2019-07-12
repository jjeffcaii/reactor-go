package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

type fluxCreate struct {
	source       func(context.Context, Sink)
	backpressure OverflowStrategy
}

func (m fluxCreate) SubscribeWith(ctx context.Context, s rs.Subscriber) {
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

func newFluxCreate(c func(ctx context.Context, sink Sink)) *fluxCreate {
	return &fluxCreate{
		source:       c,
		backpressure: OverflowBuffer,
	}
}
