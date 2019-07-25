package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type monoError struct {
	e error
}

func (p monoError) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	actual := internal.NewCoreSubscriber(ctx, s)
	actual.OnSubscribe(internal.EmptySubscription)
	actual.OnError(p.e)
}

func newMonoError(e error) monoError {
	return monoError{e: e}
}
