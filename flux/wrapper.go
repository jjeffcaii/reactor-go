package flux

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type wrapper struct {
	rs.RawPublisher
}

func (p wrapper) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	p.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (p wrapper) Filter(f rs.Predicate) Flux {
	return wrap(newFluxFilter(p, f))
}

func (p wrapper) Map(t rs.Transformer) Flux {
	return wrap(newFluxMap(p, t))
}

func (p wrapper) SubscribeOn(sc scheduler.Scheduler) Flux {
	return wrap(newFluxSubscribeOn(p, sc))
}

func wrap(r rs.RawPublisher) Flux {
	return wrapper{r}
}
