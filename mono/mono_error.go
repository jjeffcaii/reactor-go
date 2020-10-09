package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type monoError struct {
	inner error
}

func (e monoError) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	s.OnSubscribe(ctx, internal.EmptySubscription)
	s.OnError(e.inner)
}

func newMonoError(e error) monoError {
	return monoError{inner: e}
}
