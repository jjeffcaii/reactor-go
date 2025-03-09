package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	subscribers2 "github.com/jjeffcaii/reactor-go/internal/subscribers2"
)

type monoSwitchValueIfEmpty[T any] struct {
	source      reactor.RawPublisher[T]
	alternative T
}

func newMonoDoCreateIfEmpty[T any](source reactor.RawPublisher[T], value T) *monoSwitchValueIfEmpty[T] {
	return &monoSwitchValueIfEmpty[T]{
		source:      source,
		alternative: value,
	}
}

func (m monoSwitchValueIfEmpty[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	sub := subscribers2.NewSwitchValueIfEmptySubscriber[T](s, m.alternative)
	m.source.SubscribeWith(ctx, sub)
}
