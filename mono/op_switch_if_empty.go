package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoSwitchIfEmpty struct {
	*baseMono
	source Mono
	other  Mono
}

func (m *monoSwitchIfEmpty) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (m *monoSwitchIfEmpty) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	s := subscribers.NewSwitchIfEmptySubscriber(m.other, actual)
	actual.OnSubscribe(s)
	m.source.SubscribeWith(ctx, s)
}

func newMonoSwitchIfEmpty(source, other Mono) Mono {
	m := &monoSwitchIfEmpty{
		source: source,
		other:  other,
	}
	m.baseMono = &baseMono{
		child: m,
	}
	return m
}
