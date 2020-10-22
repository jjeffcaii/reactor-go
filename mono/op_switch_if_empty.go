package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoSwitchIfEmpty struct {
	source reactor.RawPublisher
	other  reactor.RawPublisher
}

func (m *monoSwitchIfEmpty) Parent() reactor.RawPublisher {
	return m.source
}

func (m *monoSwitchIfEmpty) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	s := subscribers.NewSwitchIfEmptySubscriber(m.other, actual)
	actual.OnSubscribe(ctx, s)
	m.source.SubscribeWith(ctx, s)
}

func newMonoSwitchIfEmpty(source, other reactor.RawPublisher) *monoSwitchIfEmpty {
	return &monoSwitchIfEmpty{
		source: source,
		other:  other,
	}
}
