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

func (p *fluxSubscribeOn) Subscribe(ctx context.Context, s rs.Subscriber) {
	w := p.sc.Worker()
	w.Do(func() {
		defer func() {
			_ = w.Close()
		}()
		p.source.Subscribe(ctx, s)
	})
}

func (p *fluxSubscribeOn) Filter(filter rs.Predicate) Flux {
	return newFluxFilter(p, filter)
}

func (p *fluxSubscribeOn) Map(t rs.Transformer) Flux {
	return newFluxMap(p, t)
}

func (p *fluxSubscribeOn) SubscribeOn(sc scheduler.Scheduler) Flux {
	p.sc = sc
	return p
}

func newFluxSubscribeOn(source Flux, sc scheduler.Scheduler) Flux {
	return &fluxSubscribeOn{
		source: source,
		sc:     sc,
	}
}
