package mono

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

var pool *sync.Pool = &sync.Pool{
	New: func() interface{} {
		return wrap(nil)
	},
}

var errNotProcessor = errors.New("publisher is not a Processor")

type wrapper struct {
	reactor.RawPublisher
}

func (p wrapper) Subscribe(ctx context.Context, options ...reactor.SubscriberOption) {
	p.SubscribeWith(ctx, reactor.NewSubscriber(options...))
	putObjToPool(&p)
}

func (p wrapper) SwitchIfEmpty(alternative Mono) Mono {
	var w = getObjFromPool(newMonoSwitchIfEmpty(p.RawPublisher, alternative))
	putObjToPool(&p)
	return w
}

func (p wrapper) Filter(f reactor.Predicate) Mono {
	var w = getObjFromPool(newMonoFilter(p.RawPublisher, f))
	putObjToPool(&p)
	return w
}

func (p wrapper) Map(t reactor.Transformer) Mono {
	var w = getObjFromPool(newMonoMap(p.RawPublisher, t))
	putObjToPool(&p)
	return w

}

func (p wrapper) FlatMap(mapper FlatMapper) Mono {
	var w = getObjFromPool(newMonoFlatMap(p.RawPublisher, mapper))
	putObjToPool(&p)
	return w
}

func (p wrapper) SubscribeOn(sc scheduler.Scheduler) Mono {
	var w = getObjFromPool(newMonoScheduleOn(p.RawPublisher, sc))
	putObjToPool(&p)
	return w
}

func (p wrapper) DoOnNext(fn reactor.FnOnNext) Mono {
	var w = getObjFromPool(newMonoPeek(p.RawPublisher, peekNext(fn)))
	putObjToPool(&p)
	return w
}

func (p wrapper) DoOnError(fn reactor.FnOnError) Mono {
	var w = getObjFromPool(newMonoPeek(p.RawPublisher, peekError(fn)))
	putObjToPool(&p)
	return w
}

func (p wrapper) DoOnComplete(fn reactor.FnOnComplete) Mono {
	var w = getObjFromPool(newMonoPeek(p.RawPublisher, peekComplete(fn)))
	putObjToPool(&p)
	return w
}

func (p wrapper) DoOnCancel(fn reactor.FnOnCancel) Mono {
	var w = getObjFromPool(newMonoPeek(p.RawPublisher, peekCancel(fn)))
	putObjToPool(&p)
	return w
}

func (p wrapper) DoOnDiscard(fn reactor.FnOnDiscard) Mono {
	var w = getObjFromPool(newMonoContext(p.RawPublisher, withContextDiscard(fn)))
	putObjToPool(&p)
	return w
}

func (p wrapper) DoFinally(fn reactor.FnOnFinally) Mono {
	var w = getObjFromPool(newMonoDoFinally(p.RawPublisher, fn))
	putObjToPool(&p)
	return w
}

func (p wrapper) DoOnSubscribe(fn reactor.FnOnSubscribe) Mono {
	var w = getObjFromPool(newMonoPeek(p.RawPublisher, peekSubscribe(fn)))
	putObjToPool(&p)
	return w
}

func (p wrapper) DelayElement(delay time.Duration) Mono {
	var w = getObjFromPool(newMonoDelayElement(p.RawPublisher, delay, scheduler.Elastic()))
	putObjToPool(&p)
	return w
}

func (p wrapper) Timeout(timeout time.Duration) Mono {
	if timeout <= 0 {
		return p
	}
	var w = getObjFromPool(newMonoTimeout(p.RawPublisher, timeout))
	putObjToPool(&p)
	return w
}

func (p wrapper) Block(ctx context.Context) (Any, error) {
	done := make(chan struct{})
	vchan := make(chan reactor.Any, 1)
	echan := make(chan error, 1)
	b := subscribers.NewBlockSubscriber(done, vchan, echan)
	p.SubscribeWith(ctx, b)
	<-done

	defer close(vchan)
	defer close(echan)
	putObjToPool(&p)
	select {
	case value := <-vchan:
		return value, nil
	case err := <-echan:
		return nil, err
	default:
		return nil, nil
	}
}

func (p wrapper) Success(v Any) {
	p.mustProcessor().Success(v)
	putObjToPool(&p)
}

func (p wrapper) Error(e error) {
	p.mustProcessor().Error(e)
	putObjToPool(&p)
}

func (p wrapper) mustProcessor() *processor {
	pp, ok := p.RawPublisher.(*processor)
	if !ok {
		panic(errNotProcessor)
	}
	putObjToPool(&p)
	return pp
}

func getObjFromPool(r reactor.RawPublisher) wrapper {
	var w = pool.Get().(wrapper)
	w.RawPublisher = r
	return w
}

func putObjToPool(w *wrapper)  {
	w.RawPublisher = nil
	pool.Put(*w)
}

func wrap(r reactor.RawPublisher) wrapper {
	return wrapper{r}
}
