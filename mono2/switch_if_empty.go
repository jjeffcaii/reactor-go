package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	subscribers2 "github.com/jjeffcaii/reactor-go/internal/subscribers2"
)

type monoSwitchIfEmpty[T any] struct {
	source reactor.RawPublisher[T]
	other  reactor.RawPublisher[T]
}

func newMonoSwitchIfEmpty[T any](origin, alternative reactor.RawPublisher[T]) *monoSwitchIfEmpty[T] {
	return &monoSwitchIfEmpty[T]{
		source: origin,
		other:  alternative,
	}
}

func (m *monoSwitchIfEmpty[T]) SubscribeWith(ctx context.Context, actual reactor.Subscriber[T]) {
	sub := subscribers2.NewSwitchIfEmptySubscriber(m.other, actual)
	actual.OnSubscribe(ctx, sub)
	m.source.SubscribeWith(ctx, sub)
}
