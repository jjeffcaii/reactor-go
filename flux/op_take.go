package flux

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
)

type fluxTake struct {
	source reactor.RawPublisher
	n      int
}

func (ft *fluxTake) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	actual := internal.ExtractRawSubscriber(s)
	take := newTakeSubscriber(actual, int64(ft.n))
	ft.source.SubscribeWith(ctx, internal.NewCoreSubscriber(take))
}

type takeSubscriber struct {
	actual    reactor.Subscriber
	remaining int64
	stat      int32
	su        reactor.Subscription
}

func (ts *takeSubscriber) OnError(e error) {
	if atomic.CompareAndSwapInt32(&ts.stat, 0, statError) {
		ts.actual.OnError(e)
		return
	}
	hooks.Global().OnErrorDrop(e)
}

func (ts *takeSubscriber) OnNext(v Any) {
	remaining := atomic.AddInt64(&ts.remaining, -1)
	// if no remaining or stat is not default value.
	if remaining < 0 || atomic.LoadInt32(&ts.stat) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	ts.actual.OnNext(v)
	if remaining > 0 {
		return
	}
	ts.su.Cancel()
	ts.OnComplete()
}

func (ts *takeSubscriber) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	if atomic.LoadInt64(&ts.remaining) < 1 {
		su.Cancel()
		return
	}
	ts.su = su
	ts.actual.OnSubscribe(ctx, su)
}

func (ts *takeSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&ts.stat, 0, statComplete) {
		ts.actual.OnComplete()
	}
}

func newTakeSubscriber(actual reactor.Subscriber, n int64) *takeSubscriber {
	return &takeSubscriber{
		actual:    actual,
		remaining: n,
	}
}

func newFluxTake(source reactor.RawPublisher, n int) *fluxTake {
	return &fluxTake{
		source: source,
		n:      n,
	}
}
