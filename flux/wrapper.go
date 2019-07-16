package flux

import (
	"context"
	"errors"

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

func (p wrapper) DoOnNext(fn rs.FnOnNext) Flux {
	return wrap(newFluxPeek(p, peekNext(fn)))
}

func (p wrapper) DoOnComplete(fn rs.FnOnComplete) Flux {
	return wrap(newFluxPeek(p, peekComplete(fn)))
}
func (p wrapper) DoOnRequest(fn rs.FnOnRequest) Flux {
	return wrap(newFluxPeek(p, peekRequest(fn)))
}

func (p wrapper) DoOnDiscard(fn rs.FnOnDiscard) Flux {
	return wrap(newFluxContext(p, withContextDiscard(fn)))
}

func (p wrapper) DoFinally(fn rs.FnOnFinally) Flux {
	return wrap(newFluxFinally(p, fn))
}

func (p wrapper) Complete() {
	p.mustProcessor().Complete()
}

func (p wrapper) Error(e error) {
	p.mustProcessor().Error(e)
}

func (p wrapper) Next(v interface{}) {
	p.mustProcessor().Next(v)
}

func (p wrapper) mustProcessor() rawProcessor {
	v, ok := p.RawPublisher.(rawProcessor)
	if !ok {
		panic(errors.New("require a processor"))
	}
	return v
}

func wrap(r rs.RawPublisher) Flux {
	return wrapper{r}
}
func wrapProcessor(r rawProcessor) Processor {
	return wrapper{r}
}
