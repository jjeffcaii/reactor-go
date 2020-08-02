package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type fluxSubscribeOn struct {
	source reactor.RawPublisher
	sc     scheduler.Scheduler
}

func (p *fluxSubscribeOn) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	p.sc.Worker().Do(func() {
		p.source.SubscribeWith(ctx, s)
	})
}

func newFluxSubscribeOn(source reactor.RawPublisher, sc scheduler.Scheduler) *fluxSubscribeOn {
	return &fluxSubscribeOn{
		source: source,
		sc:     sc,
	}
}
