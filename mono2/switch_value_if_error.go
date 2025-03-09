package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	subscribers2 "github.com/jjeffcaii/reactor-go/internal/subscribers2"
)

type monoSwitchValueIfError[T any] struct {
	source      reactor.RawPublisher[T]
	alternative T
}

func newMonoDoCreateIfError[T any](source reactor.RawPublisher[T], value T) *monoSwitchValueIfError[T] {
	return &monoSwitchValueIfError[T]{
		source:      source,
		alternative: value,
	}
}

func (m monoSwitchValueIfError[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	sub := subscribers2.NewSwitchValueIfErrorSubscriber[T](s, m.alternative)
	m.source.SubscribeWith(ctx, sub)
}
