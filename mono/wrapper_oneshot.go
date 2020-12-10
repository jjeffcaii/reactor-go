package mono

import (
	"context"
	"sync"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

var _oneshotWrapperPool = sync.Pool{
	New: func() interface{} {
		return new(oneshotWrapper)
	},
}

type oneshotWrapper struct {
	reactor.RawPublisher
}

func borrowOneshotWrapper(origin reactor.RawPublisher) *oneshotWrapper {
	wrapper := _oneshotWrapperPool.Get().(*oneshotWrapper)
	wrapper.RawPublisher = origin
	return wrapper
}

func returnOneshotWrapper(o *oneshotWrapper) (raw reactor.RawPublisher) {
	raw, o.RawPublisher = o.RawPublisher, nil
	_oneshotWrapperPool.Put(o)
	return
}

func (o *oneshotWrapper) Subscribe(ctx context.Context, options ...reactor.SubscriberOption) {
	returnOneshotWrapper(o).SubscribeWith(ctx, reactor.NewSubscriber(options...))
}

func (o *oneshotWrapper) Filter(predicate reactor.Predicate) Mono {
	o.RawPublisher = newMonoFilter(o.RawPublisher, predicate)
	return o
}

func (o *oneshotWrapper) Map(transformer reactor.Transformer) Mono {
	o.RawPublisher = newMonoMap(o.RawPublisher, transformer)
	return o
}

func (o *oneshotWrapper) FlatMap(mapper FlatMapper) Mono {
	o.RawPublisher = newMonoFlatMap(o.RawPublisher, mapper)
	return o
}

func (o *oneshotWrapper) SubscribeOn(scheduler scheduler.Scheduler) Mono {
	o.RawPublisher = newMonoScheduleOn(o.RawPublisher, scheduler)
	return o
}

func (o *oneshotWrapper) Block(ctx context.Context) (Any, error) {
	return block(ctx, returnOneshotWrapper(o))
}

func (o *oneshotWrapper) DoOnNext(next reactor.FnOnNext) Mono {
	o.RawPublisher = newMonoPeek(o.RawPublisher, peekNext(next))
	return o
}

func (o *oneshotWrapper) DoOnComplete(complete reactor.FnOnComplete) Mono {
	o.RawPublisher = newMonoPeek(o.RawPublisher, peekComplete(complete))
	return o
}

func (o *oneshotWrapper) DoOnSubscribe(subscribe reactor.FnOnSubscribe) Mono {
	o.RawPublisher = newMonoPeek(o.RawPublisher, peekSubscribe(subscribe))
	return o
}

func (o *oneshotWrapper) DoOnError(onError reactor.FnOnError) Mono {
	o.RawPublisher = newMonoPeek(o.RawPublisher, peekError(onError))
	return o
}

func (o *oneshotWrapper) DoOnCancel(cancel reactor.FnOnCancel) Mono {
	o.RawPublisher = newMonoPeek(o.RawPublisher, peekCancel(cancel))
	return o
}

func (o *oneshotWrapper) DoFinally(finally reactor.FnOnFinally) Mono {
	o.RawPublisher = newMonoDoFinally(o.RawPublisher, finally)
	return o
}
func (o *oneshotWrapper) SwitchValueIfError(v Any) Mono {
	o.RawPublisher = newMonoDoCreateIfError(o.RawPublisher, v)
	return o
}

func (o *oneshotWrapper) DoOnDiscard(discard reactor.FnOnDiscard) Mono {
	o.RawPublisher = newMonoContext(o.RawPublisher, withContextDiscard(discard))
	return o
}

func (o *oneshotWrapper) SwitchIfEmpty(alternative Mono) Mono {
	o.RawPublisher = newMonoSwitchIfEmpty(o.RawPublisher, alternative)
	return o
}

func (o *oneshotWrapper) SwitchIfError(alternative func(error) Mono) Mono {
	o.RawPublisher = newMonoSwitchIfError(o.RawPublisher, alternative)
	return o
}

func (o *oneshotWrapper) DelayElement(delay time.Duration) Mono {
	o.RawPublisher = newMonoDelayElement(o.RawPublisher, delay, scheduler.Parallel())
	return o
}

func (o *oneshotWrapper) Timeout(timeout time.Duration) Mono {
	o.RawPublisher = newMonoTimeout(o.RawPublisher, timeout)
	return o
}

func (o *oneshotWrapper) ZipWith(other Mono) Mono {
	return o.ZipCombineWith(other, nil)
}

func (o *oneshotWrapper) ZipCombineWith(other Mono, cmb Combinator) Mono {
	second := unpackRawPublisher(other)
	pubs := []reactor.RawPublisher{
		o.RawPublisher,
		second,
	}
	o.RawPublisher = newMonoZip(pubs, cmb)
	return o
}

func (o *oneshotWrapper) Raw() reactor.RawPublisher {
	return o.RawPublisher
}
