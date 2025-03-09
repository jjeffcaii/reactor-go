package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

func Empty[T any]() Mono[T] {
	return wrap[T](newMonoEmpty[T]())
}

type emptySubscription[T any] struct {
}

func (emptySubscription[T]) Request(_ context.Context, _ int) {
}

func (emptySubscription[T]) Cancel(_ context.Context) {
}

type monoEmpty[T any] struct {
}

func newMonoEmpty[T any]() monoEmpty[T] {
	return monoEmpty[T]{}
}

func (m monoEmpty[T]) SubscribeWith(ctx context.Context, actual reactor.Subscriber[T]) {
	var su emptySubscription[T]
	actual.OnSubscribe(ctx, su)
	actual.OnComplete(ctx)
}
