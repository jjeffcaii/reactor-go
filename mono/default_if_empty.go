package mono

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

var (
	_ reactor.Subscriber   = (*defaultIfEmptySubscriber)(nil)
	_ reactor.Subscription = (*defaultIfEmptySubscriber)(nil)
)

type monoDefaultIfEmpty struct {
	source       reactor.RawPublisher
	defaultValue reactor.Any
}

func newMonoDefaultIfEmpty(source reactor.RawPublisher, defaultValue reactor.Any) *monoDefaultIfEmpty {
	return &monoDefaultIfEmpty{
		source:       source,
		defaultValue: defaultValue,
	}
}

func (m *monoDefaultIfEmpty) SubscribeWith(ctx context.Context, subscriber reactor.Subscriber) {
	s := &defaultIfEmptySubscriber{
		actual:       subscriber,
		defaultValue: m.defaultValue,
	}
	m.source.SubscribeWith(ctx, s)
}

type defaultIfEmptySubscriber struct {
	actual       reactor.Subscriber
	cntEmit      int32
	defaultValue reactor.Any
	su           reactor.Subscription
}

func (d *defaultIfEmptySubscriber) OnComplete() {
	if atomic.LoadInt32(&d.cntEmit) > 0 || d.defaultValue == nil {
		d.actual.OnComplete()
		return
	}
	d.OnNext(d.defaultValue)
	d.actual.OnComplete()
}

func (d *defaultIfEmptySubscriber) OnError(err error) {
	if !atomic.CompareAndSwapInt32(&d.cntEmit, 0, 1) {
		hooks.Global().OnErrorDrop(err)
		return
	}
	d.actual.OnError(err)
}

func (d *defaultIfEmptySubscriber) OnNext(any reactor.Any) {
	if !atomic.CompareAndSwapInt32(&d.cntEmit, 0, 1) {
		hooks.Global().OnNextDrop(any)
		return
	}
	d.actual.OnNext(any)
}

func (d *defaultIfEmptySubscriber) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	d.su = su
	d.actual.OnSubscribe(ctx, su)
}

func (d *defaultIfEmptySubscriber) Request(n int) {
	d.su.Request(n)
}

func (d *defaultIfEmptySubscriber) Cancel() {
	d.su.Cancel()
}
