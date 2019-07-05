package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type fluxSubscribeOn struct {
	source Flux
	sc     scheduler.Scheduler
}

func (p *fluxSubscribeOn) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	w := p.sc.Worker()
	w.Do(func() {
		defer func() {
			_ = w.Close()
		}()
		p.source.SubscribeWith(ctx, s)
	})
}

func newFluxSubscribeOn(source Flux, sc scheduler.Scheduler) *fluxSubscribeOn {
	return &fluxSubscribeOn{
		source: source,
		sc:     sc,
	}
}
