package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoSwitchIfError struct {
	source reactor.RawPublisher
	sw     func(error) Mono
}

func (m monoSwitchIfError) Parent() reactor.RawPublisher {
	return m.source
}

func (m monoSwitchIfError) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	alternative := func(err error) reactor.RawPublisher {
		return m.sw(err)
	}
	s := subscribers.NewSwitchIfErrorSubscriber(alternative, actual)
	actual.OnSubscribe(ctx, s)
	m.source.SubscribeWith(ctx, s)
}

func newMonoSwitchIfError(source reactor.RawPublisher, sw func(error) Mono) monoSwitchIfError {
	return monoSwitchIfError{
		source: source,
		sw:     sw,
	}
}
