package flux

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type wrapper struct {
	raw
}

func (p wrapper) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	p.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (p wrapper) Filter(f rs.Predicate) Flux {
	return wrapper{newFluxFilter(p, f)}
}

func (p wrapper) Map(t rs.Transformer) Flux {
	return wrapper{newFluxMap(p, t)}
}

func (p wrapper) SubscribeOn(sc scheduler.Scheduler) Flux {
	return wrapper{newFluxSubscribeOn(p, sc)}
}
