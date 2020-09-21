package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoSwitchIfEmpty struct {
	source reactor.RawPublisher
	other  reactor.RawPublisher
}

func (m *monoSwitchIfEmpty) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
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
