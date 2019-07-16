package flux

import (
	"context"
	"os"
	"strconv"

	"github.com/jjeffcaii/reactor-go"
)

var (
	BuffSizeXS = 32
	BuffSizeSM = 256
)

func init() {
	if sm, ok := os.LookupEnv("REACTOR_BUFF_SM"); ok {
		v, err := strconv.Atoi(sm)
		if err != nil {
			panic(err)
		}
		if v < 16 {
			v = 16
		}
		BuffSizeSM = v
	}

	if xs, ok := os.LookupEnv("REACTOR_BUFF_XS"); ok {
		v, err := strconv.Atoi(xs)
		if err != nil {
			panic(err)
		}
		if v < 8 {
			v = 8
		}
		BuffSizeXS = v
	}
}

type fluxCreate struct {
	source       func(context.Context, Sink)
	backpressure OverflowStrategy
}

func (m fluxCreate) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	var sink interface {
		rs.Subscription
		Sink
	}
	switch m.backpressure {
	case OverflowDrop:
		// TODO: need implementation
		panic("implement me")
	case OverflowError:
		// TODO: need implementation
		panic("implement me")
	case OverflowIgnore:
		// TODO: need implementation
		panic("implement me")
	case OverflowLatest:
		// TODO: need implementation
		panic("implement me")
	default:
		sink = newBufferedSink(s, BuffSizeSM)
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
