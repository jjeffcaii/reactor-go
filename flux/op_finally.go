package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type fluxFinally struct {
	source    reactor.RawPublisher
	onFinally reactor.FnOnFinally
}

func (p *fluxFinally) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(subscribers.NewDoFinallySubscriber(actual, p.onFinally))
	p.source.SubscribeWith(ctx, actual)
}

func newFluxFinally(source reactor.RawPublisher, onFinally reactor.FnOnFinally) *fluxFinally {
	return &fluxFinally{
		source:    source,
		onFinally: onFinally,
	}
}
