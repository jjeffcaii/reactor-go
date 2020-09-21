package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoDoFinally struct {
	source    reactor.RawPublisher
	onFinally reactor.FnOnFinally
}

func (m *monoDoFinally) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(subscribers.NewDoFinallySubscriber(actual, m.onFinally))
	m.source.SubscribeWith(ctx, actual)
}

func newMonoDoFinally(source reactor.RawPublisher, onFinally reactor.FnOnFinally) *monoDoFinally {
	return &monoDoFinally{
		source:    source,
		onFinally: onFinally,
	}
}
