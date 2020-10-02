package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoDoFinally struct {
	source    reactor.RawPublisher
	onFinally reactor.FnOnFinally
}

func (m *monoDoFinally) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	m.source.SubscribeWith(ctx, subscribers.NewDoFinallySubscriber(s, m.onFinally))
}

func newMonoDoFinally(source reactor.RawPublisher, onFinally reactor.FnOnFinally) *monoDoFinally {
	return &monoDoFinally{
		source:    source,
		onFinally: onFinally,
	}
}
