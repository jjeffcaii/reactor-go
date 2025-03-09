package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	subscribers2 "github.com/jjeffcaii/reactor-go/internal/subscribers2"
)

type fluxFinally[T any] struct {
	origin    reactor.RawPublisher[T]
	onFinally reactor.FnOnFinally
}

func newFluxFinally[T any](origin reactor.RawPublisher[T], onFinally reactor.FnOnFinally) *fluxFinally[T] {
	return &fluxFinally[T]{
		origin:    origin,
		onFinally: onFinally,
	}
}

func (p *fluxFinally[T]) SubscribeWith(ctx context.Context, actual reactor.Subscriber[T]) {
	p.origin.SubscribeWith(ctx, subscribers2.NewDoFinallySubscriber[T](actual, p.onFinally))
}
