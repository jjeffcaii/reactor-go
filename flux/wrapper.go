package flux

import (
	"context"
	"errors"

	reactor "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type wrapper struct {
	reactor.RawPublisher
}

func (p wrapper) Subscribe(ctx context.Context, options ...reactor.SubscriberOption) {
	p.SubscribeWith(ctx, reactor.NewSubscriber(options...))
}

func (p wrapper) PublishOn(sc scheduler.Scheduler) Flux {
	return wrap(newFluxPublishOn(p.RawPublisher, sc, reactor.RequestInfinite, reactor.RequestInfinite, BuffSizeSM))
}

func (p wrapper) RateLimit(prefetch, lowTide int) Flux {
	return wrap(newFluxPublishOn(p.RawPublisher, scheduler.Elastic(), prefetch, lowTide, BuffSizeSM))
}

func (p wrapper) Take(n int) Flux {
	return wrap(newFluxTake(p.RawPublisher, n))
}

func (p wrapper) Filter(f reactor.Predicate) Flux {
	return wrap(newFluxFilter(p.RawPublisher, f))
}

func (p wrapper) Map(t reactor.Transformer) Flux {
	return wrap(newFluxMap(p.RawPublisher, t))
}

func (p wrapper) SubscribeOn(sc scheduler.Scheduler) Flux {
	return wrap(newFluxSubscribeOn(p.RawPublisher, sc))
}

func (p wrapper) DoOnNext(fn reactor.FnOnNext) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekNext(fn)))
}

func (p wrapper) DoOnComplete(fn reactor.FnOnComplete) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekComplete(fn)))
}
func (p wrapper) DoOnRequest(fn reactor.FnOnRequest) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekRequest(fn)))
}

func (p wrapper) DoOnDiscard(fn reactor.FnOnDiscard) Flux {
	return wrap(newFluxContext(p.RawPublisher, withContextDiscard(fn)))
}

func (p wrapper) DoOnCancel(fn reactor.FnOnCancel) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekCancel(fn)))
}

func (p wrapper) DoOnError(fn reactor.FnOnError) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekError(fn)))
}

func (p wrapper) DoFinally(fn reactor.FnOnFinally) Flux {
	return wrap(newFluxFinally(p.RawPublisher, fn))
}

func (p wrapper) DoOnSubscribe(fn reactor.FnOnSubscribe) Flux {
	return wrap(newFluxPeek(p.RawPublisher, peekSubscribe(fn)))
}

func (p wrapper) SwitchOnFirst(fn FnSwitchOnFirst) Flux {
	return wrap(newFluxSwitchOnFirst(p.RawPublisher, fn))
}

func (p wrapper) ToChan(ctx context.Context, cap int) (<-chan interface{}, <-chan error) {
	if cap < 1 {
		cap = 1
	}
	ch := make(chan interface{}, cap)
	err := make(chan error, 1)
	p.
		DoFinally(func(s reactor.SignalType) {
			if s == reactor.SignalTypeCancel {
				err <- reactor.ErrSubscribeCancelled
			}
			close(ch)
			close(err)
		}).
		SubscribeOn(scheduler.Elastic()).
		Subscribe(ctx,
			reactor.OnNext(func(v interface{}) {
				ch <- v
			}),
			reactor.OnError(func(e error) {
				err <- e
			}),
		)
	return ch, err
}

func (p wrapper) BlockFirst(ctx context.Context) (first interface{}, err error) {
	done := make(chan struct{})
	var su reactor.Subscription
	p.
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		Subscribe(ctx, reactor.OnNext(func(v interface{}) {
			first = v
			su.Cancel()
		}), reactor.OnSubscribe(func(s reactor.Subscription) {
			su = s
			su.Request(1)
		}), reactor.OnError(func(e error) {
			err = e
		}))
	<-done
	return
}

func (p wrapper) BlockLast(ctx context.Context) (last interface{}, err error) {
	done := make(chan struct{})
	p.
		DoFinally(func(s reactor.SignalType) {
			close(done)
		}).
		DoOnCancel(func() {
			err = reactor.ErrSubscribeCancelled
		}).
		Subscribe(ctx,
			reactor.OnNext(func(v interface{}) {
				if old := last; old != nil {
					hooks.Global().OnNextDrop(old)
				}
				last = v
			}),
			reactor.OnError(func(e error) {
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

func wrap(r reactor.RawPublisher) wrapper {
	return wrapper{r}
}
