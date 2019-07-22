package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoSwitchIfEmpty struct {
	source rs.RawPublisher
	other  rs.RawPublisher
}

func (m *monoSwitchIfEmpty) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	s := subscribers.NewSwitchIfEmptySubscriber(m.other, actual)
	actual.OnSubscribe(s)
	m.source.SubscribeWith(ctx, s)
}

func newMonoSwitchIfEmpty(source, other rs.RawPublisher) *monoSwitchIfEmpty {
	return &monoSwitchIfEmpty{
		source: source,
		other:  other,
	}
}
