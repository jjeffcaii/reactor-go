package mono

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
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

func (p *delayElementSubscriber) Request(n int) {
	p.s.Request(n)
}

func (p *delayElementSubscriber) Cancel() {
	// TODO: support cancel
	p.s.Cancel()
}

func (p *delayElementSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&(p.stat), 0, statError) {
		p.actual.OnError(err)
		return
	}
	hooks.Global().OnErrorDrop(err)
}

func (p *delayElementSubscriber) OnNext(v Any) {
	if !atomic.CompareAndSwapInt32(&(p.stat), 0, statComplete) {
		hooks.Global().OnNextDrop(v)
		return
	}
	p.v = v
	time.AfterFunc(p.delay, func() {
		p.actual.OnNext(v)
	})
}

func (p *delayElementSubscriber) OnSubscribe(s reactor.Subscription) {
	p.s = s
	p.actual.OnSubscribe(p)
	s.Request(reactor.RequestInfinite)
}

func (p *delayElementSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&(p.stat), 0, statComplete) {
		p.actual.OnComplete()
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
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, newDelayElementSubscriber(actual, p.delay, p.sc))
	p.source.SubscribeWith(ctx, actual)
}

func newMonoDelayElement(source reactor.RawPublisher, delay time.Duration, sc scheduler.Scheduler) *monoDelayElement {
	return &monoDelayElement{
		source: source,
		delay:  delay,
		sc:     sc,
	}
}
