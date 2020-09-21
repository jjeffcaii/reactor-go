package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type monoError struct {
	e error
}

func (p monoError) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	actual := internal.NewCoreSubscriber(s)
	actual.OnSubscribe(ctx, internal.EmptySubscription)
	actual.OnError(p.e)
}

func newMonoError(e error) monoError {
	return monoError{e: e}
}
