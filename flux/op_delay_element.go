package flux

import (
	"context"
	"sync/atomic"
	"time"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type delayElementSubscriber struct {
	delay  time.Duration
	sc     scheduler.Scheduler
	s      rs.Subscription
	actual rs.Subscriber
	stat   int32
}

func (d *delayElementSubscriber) Request(n int) {
	d.s.Request(n)
}

func (d *delayElementSubscriber) Cancel() {
	// TODO: support cancel
	d.s.Cancel()
}

func (d *delayElementSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&d.stat, 0, statComplete) {
		d.actual.OnComplete()
	}
}

func (d *delayElementSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&d.stat, 0, statError) {
		d.actual.OnError(err)
		return
	}
	hooks.Global().OnErrorDrop(err)
}

func (d *delayElementSubscriber) OnNext(v interface{}) {
	if atomic.LoadInt32(&d.stat) == statComplete {
		hooks.Global().OnNextDrop(v)
		return
	}
	<-time.After(d.delay)
	d.actual.OnNext(v)
}

func (d *delayElementSubscriber) OnSubscribe(s rs.Subscription) {
	d.s = s
	d.actual.OnSubscribe(d)
	s.Request(rs.RequestInfinite)
}

func newDelayElementSubscriber(actual rs.Subscriber, delay time.Duration, sc scheduler.Scheduler) rs.Subscriber {
	return &delayElementSubscriber{
		delay:  delay,
		actual: actual,
		sc:     sc,
	}
}

type fluxDelayElement struct {
	source rs.RawPublisher
	delay  time.Duration
	sc     scheduler.Scheduler
}

func (f *fluxDelayElement) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, newDelayElementSubscriber(actual, f.delay, f.sc))
	f.source.SubscribeWith(ctx, actual)
}

func newFluxDelayElement(source rs.RawPublisher, delay time.Duration, sc scheduler.Scheduler) *fluxDelayElement {
	return &fluxDelayElement{
		source: source,
		delay:  delay,
		sc:     sc,
	}
}
