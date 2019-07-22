package flux

import (
	"context"
	"errors"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type wrapper struct {
	rs.RawPublisher
}

func (p wrapper) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	p.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (p wrapper) Filter(f rs.Predicate) Flux {
	return wrap(newFluxFilter(p.RawPublisher, f))
}

func (p wrapper) Map(t rs.Transformer) Flux {
	return wrap(newFluxMap(p.RawPublisher, t))
}

func (p wrapper) SubscribeOn(sc scheduler.Scheduler) Flux {
	return wrap(newFluxSubscribeOn(p.RawPublisher, sc))
}

func (p wrapper) DoOnNext(fn rs.FnOnNext) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekNext(fn)))
}

func (p wrapper) DoOnComplete(fn rs.FnOnComplete) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekComplete(fn)))
}
func (p wrapper) DoOnRequest(fn rs.FnOnRequest) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekRequest(fn)))
}

func (p wrapper) DoOnDiscard(fn rs.FnOnDiscard) Flux {
	return wrap(newFluxContext(p.RawPublisher, withContextDiscard(fn)))
}

func (p wrapper) DoOnCancel(fn rs.FnOnCancel) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekCancel(fn)))
}

func (p wrapper) DoOnError(fn rs.FnOnError) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekError(fn)))
}

func (p wrapper) DoFinally(fn rs.FnOnFinally) Flux {
	return wrap(newFluxFinally(p.RawPublisher, fn))
}

func (p wrapper) DoOnSubscribe(fn rs.FnOnSubscribe) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekSubscribe(fn)))
}

func (p wrapper) SwitchOnFirst(fn FnSwitchOnFirst) Flux {
	return wrap(newFluxSwitchOnFirst(p.RawPublisher, fn))
}

func (p wrapper) BlockLast(ctx context.Context) (last interface{}, err error) {
	done := make(chan struct{})
	p.
		DoFinally(func(s rs.SignalType) {
			close(done)
		}).
		DoOnCancel(func() {
			err = rs.ErrSubscribeCancelled
		}).
		Subscribe(ctx,
			rs.OnNext(func(v interface{}) {
				if old := last; old != nil {
					hooks.Global().OnNextDrop(old)
				}
				last = v
			}),
			rs.OnError(func(e error) {
				err = e
			}),
		)
	<-done
	return
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

func wrap(r rs.RawPublisher) wrapper {
	return wrapper{r}
}
