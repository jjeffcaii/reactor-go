package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type wrapper struct {
	reactor.RawPublisher
}

func (w wrapper) Subscribe(ctx context.Context, options ...reactor.SubscriberOption) {
	w.SubscribeWith(ctx, reactor.NewSubscriber(options...))
}

func (w wrapper) SwitchIfEmpty(alternative Mono) Mono {
	return wrap(newMonoSwitchIfEmpty(w.RawPublisher, unpackRawPublisher(alternative)))
}

func (w wrapper) SwitchIfError(alternativeFunc func(error) Mono) Mono {
	return wrap(newMonoSwitchIfError(w.RawPublisher, alternativeFunc))
}

func (w wrapper) Filter(f reactor.Predicate) Mono {
	return wrap(newMonoFilter(w.RawPublisher, f))
}

func (w wrapper) Map(t reactor.Transformer) Mono {
	return wrap(newMonoMap(w.RawPublisher, t))
}

func (w wrapper) FlatMap(mapper FlatMapper) Mono {
	return wrap(newMonoFlatMap(w.RawPublisher, mapper))
}

func (w wrapper) SubscribeOn(sc scheduler.Scheduler) Mono {
	return wrap(newMonoScheduleOn(w.RawPublisher, sc))
}

func (w wrapper) DoOnNext(fn reactor.FnOnNext) Mono {
	return wrap(newMonoPeek(w.RawPublisher, peekNext(fn)))
}

func (w wrapper) DoOnError(fn reactor.FnOnError) Mono {
	return wrap(newMonoPeek(w.RawPublisher, peekError(fn)))
}

func (w wrapper) DoOnComplete(fn reactor.FnOnComplete) Mono {
	return wrap(newMonoPeek(w.RawPublisher, peekComplete(fn)))
}

func (w wrapper) DoOnCancel(fn reactor.FnOnCancel) Mono {
	return wrap(newMonoPeek(w.RawPublisher, peekCancel(fn)))
}

func (w wrapper) DoOnDiscard(fn reactor.FnOnDiscard) Mono {
	return wrap(newMonoContext(w.RawPublisher, withContextDiscard(fn)))
}

func (w wrapper) DoFinally(fn reactor.FnOnFinally) Mono {
	return wrap(newMonoDoFinally(w.RawPublisher, fn))
}

func (w wrapper) SwitchValueIfError(v Any) Mono {
	return wrap(newMonoDoCreateIfError(w.RawPublisher, v))
}

func (w wrapper) DoOnSubscribe(fn reactor.FnOnSubscribe) Mono {
	return wrap(newMonoPeek(w.RawPublisher, peekSubscribe(fn)))
}

func (w wrapper) DelayElement(delay time.Duration) Mono {
	return wrap(newMonoDelayElement(w.RawPublisher, delay, scheduler.Parallel()))
}

func (w wrapper) Timeout(timeout time.Duration) Mono {
	if timeout <= 0 {
		return w
	}
	return wrap(newMonoTimeout(w.RawPublisher, timeout))
}

func (w wrapper) Block(ctx context.Context) (Any, error) {
	return block(ctx, w.RawPublisher)
}

func (w wrapper) ZipWith(other Mono) Mono {
	return w.ZipCombineWith(other, nil)
}

func (w wrapper) ZipCombineWith(other Mono, cmb Combinator) Mono {
	publishers := []reactor.RawPublisher{
		w.RawPublisher,
		unpackRawPublisher(other),
	}
	return wrap(newMonoZip(publishers, cmb))
}

func (w wrapper) Raw() reactor.RawPublisher {
	return w.RawPublisher
}

func wrap(r reactor.RawPublisher) wrapper {
	return wrapper{r}
}
