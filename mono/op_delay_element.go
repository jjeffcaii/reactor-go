package mono

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type delayElementSubscriber struct {
	delay  time.Duration
	sc     scheduler.Scheduler
	s      reactor.Subscription
	actual reactor.Subscriber
	stat   int32
	v      Any
}

func (d *delayElementSubscriber) Request(n int) {
	d.s.Request(n)
}

func (d *delayElementSubscriber) Cancel() {
	// TODO: support cancel
	d.s.Cancel()
}

func (d *delayElementSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&d.stat, 0, statError) {
		d.actual.OnError(err)
		return
	}
	hooks.Global().OnErrorDrop(err)
}

func (d *delayElementSubscriber) OnNext(v Any) {
	if !atomic.CompareAndSwapInt32(&d.stat, 0, statComplete) {
		hooks.Global().OnNextDrop(v)
		return
	}
	d.v = v
	time.AfterFunc(d.delay, func() {
		d.actual.OnNext(v)
	})
}

func (d *delayElementSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	d.s = s
	d.actual.OnSubscribe(ctx, d)
	s.Request(reactor.RequestInfinite)
}

func (d *delayElementSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&d.stat, 0, statComplete) {
		d.actual.OnComplete()
	}
}

func newDelayElementSubscriber(actual reactor.Subscriber, delay time.Duration, sc scheduler.Scheduler) reactor.Subscriber {
	return &delayElementSubscriber{
		delay:  delay,
		actual: actual,
		sc:     sc,
	}
}

type monoDelayElement struct {
	source reactor.RawPublisher
	delay  time.Duration
	sc     scheduler.Scheduler
}

func (p *monoDelayElement) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	actual = newDelayElementSubscriber(actual, p.delay, p.sc)
	p.source.SubscribeWith(ctx, actual)
}

func newMonoDelayElement(source reactor.RawPublisher, delay time.Duration, sc scheduler.Scheduler) *monoDelayElement {
	return &monoDelayElement{
		source: source,
		delay:  delay,
		sc:     sc,
	}
}
