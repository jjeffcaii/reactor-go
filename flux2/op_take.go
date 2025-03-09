package flux

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

type fluxTake[T any] struct {
	source reactor.RawPublisher[T]
	n      int
}

func newFluxTake[T any](source reactor.RawPublisher[T], n int) *fluxTake[T] {
	return &fluxTake[T]{
		source: source,
		n:      n,
	}
}

func (ft *fluxTake[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	ft.source.SubscribeWith(ctx, newTakeSubscriber[T](s, int64(ft.n)))
}

type takeSubscriber[T any] struct {
	actual    reactor.Subscriber[T]
	remaining int64
	stat      int32
	su        reactor.Subscription
}

func newTakeSubscriber[T any](actual reactor.Subscriber[T], n int64) *takeSubscriber[T] {
	return &takeSubscriber[T]{
		actual:    actual,
		remaining: n,
	}
}

func (ts *takeSubscriber[T]) OnError(ctx context.Context, e error) {
	if atomic.CompareAndSwapInt32(&ts.stat, 0, statError) {
		ts.actual.OnError(ctx, e)
		return
	}
	// TODO: hooks
	//hooks.Global().OnErrorDrop(e)
}

func (ts *takeSubscriber[T]) OnNext(ctx context.Context, v T) {
	remaining := atomic.AddInt64(&ts.remaining, -1)
	// if no remaining or stat is not default value.
	if remaining < 0 || atomic.LoadInt32(&ts.stat) != 0 {
		// TODO: hooks
		//hooks.Global().OnNextDrop(v)
		return
	}
	ts.actual.OnNext(ctx, v)
	if remaining > 0 {
		return
	}
	ts.su.Cancel(ctx)
	ts.OnComplete(ctx)
}

func (ts *takeSubscriber[T]) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	if atomic.LoadInt64(&ts.remaining) < 1 {
		su.Cancel(ctx)
		return
	}
	ts.su = su
	ts.actual.OnSubscribe(ctx, su)
}

func (ts *takeSubscriber[T]) OnComplete(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&ts.stat, 0, statComplete) {
		ts.actual.OnComplete(ctx)
	}
}
