package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
	"go.uber.org/atomic"
)

type fluxTake struct {
	source rs.RawPublisher
	n      int
}

func (p *fluxTake) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	actual := internal.ExtractRawSubscriber(s)
	take := &takeSubscriber{
		actual:    actual,
		remaining: atomic.NewInt64(int64(p.n)),
		stat:      atomic.NewInt32(0),
	}
	p.source.SubscribeWith(ctx, internal.NewCoreSubscriber(ctx, take))
}

type takeSubscriber struct {
	actual    rs.Subscriber
	remaining *atomic.Int64
	stat      *atomic.Int32
	su        rs.Subscription
}

func (t *takeSubscriber) OnError(e error) {
	if !t.stat.CAS(0, statError) {
		hooks.Global().OnErrorDrop(e)
		return
	}
	t.actual.OnError(e)
}

func (t *takeSubscriber) OnNext(v interface{}) {
	remaining := t.remaining.Dec()
	// if no remaining or stat is not default value.
	if remaining < 0 || t.stat.Load() != 0 {
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

func (t *takeSubscriber) OnSubscribe(su rs.Subscription) {
	if t.remaining.Load() < 1 {
		su.Cancel()
		return
	}
	t.su = su
	t.actual.OnSubscribe(su)
}

func (t *takeSubscriber) OnComplete() {
	if t.stat.CAS(0, statComplete) {
		t.actual.OnComplete()
	}
}

func newFluxTake(source rs.RawPublisher, n int) *fluxTake {
	return &fluxTake{
		source: source,
		n:      n,
	}
}
