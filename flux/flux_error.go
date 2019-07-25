package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type fluxError struct {
	e error
}

func (p fluxError) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	actual := internal.NewCoreSubscriber(ctx, s)
	actual.OnSubscribe(internal.EmptySubscription)
	actual.OnError(p.e)
}

func newFluxError(e error) fluxError {
	return fluxError{e: e}
}
