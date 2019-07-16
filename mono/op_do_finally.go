package mono

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type monoDoFinally struct {
	source    Mono
	onFinally rs.FnOnFinally
}

func (m *monoDoFinally) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, subscribers.NewDoFinallySubscriber(actual, m.onFinally))
	m.source.SubscribeWith(ctx, actual)
}

func newMonoDoFinally(source Mono, onFinally rs.FnOnFinally) *monoDoFinally {
	return &monoDoFinally{
		source:    source,
		onFinally: onFinally,
	}
}
