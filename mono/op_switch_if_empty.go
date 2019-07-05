package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoSwitchIfEmpty struct {
	source Mono
	other  Mono
}

func (m *monoSwitchIfEmpty) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	s := subscribers.NewSwitchIfEmptySubscriber(m.other, actual)
	actual.OnSubscribe(s)
	m.source.SubscribeWith(ctx, s)
}

func newMonoSwitchIfEmpty(source, other Mono) *monoSwitchIfEmpty {
	return &monoSwitchIfEmpty{
		source: source,
		other:  other,
	}
}
