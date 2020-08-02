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

func (p *fluxTake) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	actual := internal.ExtractRawSubscriber(s)
	take := newTakeSubscriber(actual, int64(p.n))
	p.source.SubscribeWith(ctx, internal.NewCoreSubscriber(ctx, take))
}

type takeSubscriber struct {
	actual    reactor.Subscriber
	remaining int64
	stat      int32
	su        reactor.Subscription
}

func (t *takeSubscriber) OnError(e error) {
	if atomic.CompareAndSwapInt32(&(t.stat), 0, statError) {
		t.actual.OnError(e)
		return
	}
	hooks.Global().OnErrorDrop(e)
}

func (t *takeSubscriber) OnNext(v Any) {
	remaining := atomic.AddInt64(&(t.remaining), -1)
	// if no remaining or stat is not default value.
	if remaining < 0 || atomic.LoadInt32(&(t.stat)) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	t.actual.OnNext(v)
	if remaining > 0 {
		return
	}
	t.su.Cancel()
	t.OnComplete()
}

func (t *takeSubscriber) OnSubscribe(su reactor.Subscription) {
	if atomic.LoadInt64(&(t.remaining)) < 1 {
		su.Cancel()
		return
	}
	t.su = su
	t.actual.OnSubscribe(su)
}

func (t *takeSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&(t.stat), 0, statComplete) {
		t.actual.OnComplete()
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
