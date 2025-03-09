package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

func Error[T any](e error) Mono[T] {
	return wrap[T](newMonoError[T](e))
}

type monoError[T any] struct {
	e error
}

func newMonoError[T any](e error) monoError[T] {
	return monoError[T]{e: e}
}

func (e monoError[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	var su emptySubscription[T]
	s.OnSubscribe(ctx, su)
	s.OnError(ctx, e.e)
}
