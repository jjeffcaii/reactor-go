package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

// Error creates a Flux which must emit an error:
func Error[T any](e error) Flux[T] {
	if e == nil {
		panic("flux: an error occurred")
	}
	return wrap[T](newFluxError[T](e))
}

type fluxError[T any] struct {
	e error
}

func newFluxError[T any](e error) fluxError[T] {
	return fluxError[T]{e: e}
}

func (p fluxError[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	var su internal.EmptySubscription
	s.OnSubscribe(ctx, su)
	s.OnError(ctx, p.e)
}
