package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
)

type fluxFinally struct {
	source    Flux
	onFinally rs.FnOnFinally
}

func (p *fluxFinally) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, subscribers.NewDoFinallySubscriber(actual, p.onFinally))
	p.source.SubscribeWith(ctx, actual)
}

func newFluxFinally(source Flux, onFinally rs.FnOnFinally) *fluxFinally {
	return &fluxFinally{
		source:    source,
		onFinally: onFinally,
	}
}
