package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type monoEmpty struct {
}

func newMonoEmpty() monoEmpty {
	return monoEmpty{}
}

func (m monoEmpty) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	actual.OnSubscribe(ctx, internal.EmptySubscription)
	actual.OnComplete()
}
