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

func Create[T any](c func(ctx context.Context, sink Sink[T]), options ...CreateOption[T]) Flux[T] {
	return wrap[T](newFluxCreate[T](c, options...))
}

// InitBuffSize initialize the size of buff. (default=256)
func InitBuffSize(size int) {
	if size < 1 {
		panic(fmt.Sprintf("invalid flux buff size: %d", size))
	}
	_buffSize = size
}

type fluxCreate[T any] struct {
	source       func(context.Context, Sink[T])
	backpressure OverflowStrategy
}

func newFluxCreate[T any](c func(context.Context, Sink[T]), options ...CreateOption[T]) *fluxCreate[T] {
	fc := &fluxCreate[T]{
		source:       c,
		backpressure: OverflowBuffer,
	}
	for i := range options {
		options[i](fc)
	}
	return fc
}

func (fc fluxCreate[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	select {
	case <-ctx.Done():
		s.OnError(ctx, reactor.NewContextError(ctx.Err()))
	default:
		var sink interface {
			reactor.Subscription
			Sink[T]
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

type CreateOption[T any] func(*fluxCreate[T])

func WithOverflowStrategy[T any](o OverflowStrategy) CreateOption[T] {
	return func(create *fluxCreate[T]) {
		create.backpressure = o
	}
}
