package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type monoScheduleOn[T any] struct {
	source reactor.RawPublisher[T]
	sc     scheduler.Scheduler
}

func newMonoScheduleOn[T any](origin reactor.RawPublisher[T], sc scheduler.Scheduler) monoScheduleOn[T] {
	return monoScheduleOn[T]{
		source: origin,
		sc:     sc,
	}
}

func (m monoScheduleOn[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	if err := m.sc.Worker().Do(func() {
		m.source.SubscribeWith(ctx, s)
	}); err != nil {
		var su internal.EmptySubscription
		s.OnSubscribe(ctx, su)
		s.OnError(ctx, err)
	}
}
