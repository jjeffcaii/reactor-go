package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type fluxError struct {
	e error
}

func (p fluxError) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	s.OnSubscribe(ctx, internal.EmptySubscription)
	s.OnError(p.e)
}

func newFluxError(e error) fluxError {
	return fluxError{e: e}
}
