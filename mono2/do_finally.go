package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	subscribers2 "github.com/jjeffcaii/reactor-go/internal/subscribers2"
)

type monoDoFinally[T any] struct {
	origin    reactor.RawPublisher[T]
	onFinally reactor.FnOnFinally
}

func newMonoDoFinally[T any](origin reactor.RawPublisher[T], onFinally reactor.FnOnFinally) monoDoFinally[T] {
	return monoDoFinally[T]{
		origin:    origin,
		onFinally: onFinally,
	}
}

func (m monoDoFinally[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	m.origin.SubscribeWith(ctx, subscribers2.NewDoFinallySubscriber[T](s, m.onFinally))
}
