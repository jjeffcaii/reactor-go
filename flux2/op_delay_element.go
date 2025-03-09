package flux

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/timerpool"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type fluxDelayElement[T any] struct {
	source reactor.RawPublisher[T]
	sc     scheduler.Scheduler
	delay  time.Duration
}

func newFluxDelayElement[T any](source reactor.RawPublisher[T], delay time.Duration, sc scheduler.Scheduler) *fluxDelayElement[T] {
	return &fluxDelayElement[T]{
		source: source,
		delay:  delay,
		sc:     sc,
	}
}

func (f *fluxDelayElement[T]) SubscribeWith(ctx context.Context, actual reactor.Subscriber[T]) {
	sub := &delayElementSubscriber[T]{
		delay:  f.delay,
		actual: actual,
		sc:     f.sc,
	}
	f.source.SubscribeWith(ctx, sub)
}

type delayElementSubscriber[T any] struct {
	delay  time.Duration
	sc     scheduler.Scheduler
	s      reactor.Subscription
	actual reactor.Subscriber[T]
	stat   int32
}

func (d *delayElementSubscriber[T]) Request(ctx context.Context, n int) {
	d.s.Request(ctx, n)
}

func (d *delayElementSubscriber[T]) Cancel(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&d.stat, 0, statCancel) {
		// TODO: support cancel
		d.s.Cancel(ctx)
	}
}

func (d *delayElementSubscriber[T]) OnComplete(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&d.stat, 0, statComplete) {
		d.actual.OnComplete(ctx)
	}
}

func (d *delayElementSubscriber[T]) OnError(ctx context.Context, err error) {
	if atomic.CompareAndSwapInt32(&d.stat, 0, statError) {
		d.actual.OnError(ctx, err)
		return
	}
	// TODO: hooks
	//hooks.Global().OnErrorDrop(err)
}

func (d *delayElementSubscriber[T]) OnNext(ctx context.Context, v T) {
	if atomic.LoadInt32(&d.stat) == statComplete {
		// TODO: hooks
		//hooks.Global().OnNextDrop(v)
		return
	}
	t := timerpool.Get(d.delay)
	defer timerpool.Put(t)

	select {
	case <-ctx.Done():
		// TODO: cancel it or???
		d.Cancel(ctx)
		// TODO: hooks
		//hooks.Global().OnNextDrop(v)
	case <-t.C:
		d.actual.OnNext(ctx, v)
	}
}

func (d *delayElementSubscriber[T]) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	d.s = s
	d.actual.OnSubscribe(ctx, d)
	s.Request(ctx, reactor.RequestInfinite)
}
