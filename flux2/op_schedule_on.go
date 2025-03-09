package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type fluxSubscribeOn[T any] struct {
	origin reactor.RawPublisher[T]
	sc     scheduler.Scheduler
}

func newFluxSubscribeOn[T any](origin reactor.RawPublisher[T], sc scheduler.Scheduler) *fluxSubscribeOn[T] {
	return &fluxSubscribeOn[T]{
		origin: origin,
		sc:     sc,
	}
}

func (p *fluxSubscribeOn[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	err := p.sc.Worker().Do(func() {
		p.origin.SubscribeWith(ctx, s)
	})
	if err != nil {
		panic(err)
	}
}
