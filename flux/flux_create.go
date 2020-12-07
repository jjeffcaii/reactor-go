package flux

import (
	"context"
	"fmt"

	"github.com/jjeffcaii/reactor-go"
)

const (
	BuffSizeXS = 32
	BuffSizeSM = 256
)

var _buffSize = BuffSizeSM

// InitBuffSize initialize the size of buff. (default=256)
func InitBuffSize(size int) {
	if size < 1 {
		panic(fmt.Sprintf("invalid flux buff size: %d", size))
	}
	_buffSize = size
}

type fluxCreate struct {
	source       func(context.Context, Sink)
	backpressure OverflowStrategy
}

type CreateOption func(*fluxCreate)

func newFluxCreate(c func(ctx context.Context, sink Sink), options ...CreateOption) *fluxCreate {
	fc := &fluxCreate{
		source:       c,
		backpressure: OverflowBuffer,
	}
	for i := range options {
		options[i](fc)
	}
	return fc
}

func (fc fluxCreate) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	select {
	case <-ctx.Done():
		s.OnError(reactor.NewContextError(ctx.Err()))
	default:
		var sink interface {
			reactor.Subscription
			Sink
		}
		switch fc.backpressure {
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
			sink = newBufferedSink(s, _buffSize)
		}
		s.OnSubscribe(ctx, sink)
		fc.source(ctx, sink)
	}
}

func WithOverflowStrategy(o OverflowStrategy) CreateOption {
	return func(create *fluxCreate) {
		create.backpressure = o
	}
}
