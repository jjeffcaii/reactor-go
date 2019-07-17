package mono

import (
	"context"
	"errors"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

var errNotProcessor = errors.New("publisher is not a Processor")

type wrapper struct {
	rs.RawPublisher
}

func (p wrapper) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	p.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (p wrapper) SwitchIfEmpty(alternative Mono) Mono {
	return wrap(newMonoSwitchIfEmpty(p, alternative))
}

func (p wrapper) Filter(f rs.Predicate) Mono {
	return wrap(newMonoFilter(p, f))
}

func (p wrapper) Map(t rs.Transformer) Mono {
	return wrap(newMonoMap(p, t))
}

func (p wrapper) FlatMap(mapper flatMapper) Mono {
	return wrap(newMonoFlatMap(p, mapper))
}

func (p wrapper) SubscribeOn(sc scheduler.Scheduler) Mono {
	return wrap(newMonoScheduleOn(p, sc))
}

func (p wrapper) DoOnNext(fn rs.FnOnNext) Mono {
	return wrap(newMonoPeek(p, peekNext(fn)))
}

func (p wrapper) DoOnError(fn rs.FnOnError) Mono {
	return wrap(newMonoPeek(p, peekError(fn)))
}

func (p wrapper) DoOnComplete(fn rs.FnOnComplete) Mono {
	return wrap(newMonoPeek(p, peekComplete(fn)))
}

func (p wrapper) DoOnCancel(fn rs.FnOnCancel) Mono {
	return wrap(newMonoPeek(p, peekCancel(fn)))
}

func (p wrapper) DoOnDiscard(fn rs.FnOnDiscard) Mono {
	return wrap(newMonoContext(p, withContextDiscard(fn)))
}

func (p wrapper) DoFinally(fn rs.FnOnFinally) Mono {
	return wrap(newMonoDoFinally(p, fn))
}

func (p wrapper) DoOnSubscribe(fn rs.FnOnSubscribe) Mono {
	return wrap(newMonoPeek(p, peekSubscribe(fn)))
}

func (p wrapper) DelayElement(delay time.Duration) Mono {
	return wrap(newMonoDelayElement(p, delay, scheduler.Elastic()))
}

func (p wrapper) Block(ctx context.Context) (interface{}, error) {
	ch := make(chan interface{}, 1)
	p.
		DoFinally(func(signal rs.SignalType) {
			close(ch)
		}).
		Subscribe(ctx,
			rs.OnNext(func(v interface{}) {
				ch <- v
			}),
			rs.OnError(func(e error) {
				ch <- e
			}),
		)
	v, ok := <-ch
	if !ok {
		return nil, nil
	}
	if err, ok := v.(error); ok {
		return nil, err
	}
	return v, nil
}

func (p wrapper) Success(v interface{}) {
	p.mustProcessor().Success(v)
}

func (p wrapper) Error(e error) {
	p.mustProcessor().Error(e)
}

func (p wrapper) mustProcessor() *processor {
	pp, ok := p.RawPublisher.(*processor)
	if !ok {
		panic(errNotProcessor)
	}
	return pp
}

func wrap(r rs.RawPublisher) Mono {
	return wrapper{r}
}

func wrapProcessor(origin *processor) Processor {
	return wrapper{origin}
}
