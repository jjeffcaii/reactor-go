package mono

import (
	"context"
	"sync/atomic"
	"time"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type delayElementSubscriber struct {
	delay  time.Duration
	sc     scheduler.Scheduler
	s      rs.Subscription
	actual rs.Subscriber
	stat   int32
	v      interface{}
}

func (p *delayElementSubscriber) Request(n int) {
	p.s.Request(n)
}

func (p *delayElementSubscriber) Cancel() {
	// TODO: support cancel
	p.s.Cancel()
}

func (p *delayElementSubscriber) OnError(err error) {
	if !atomic.CompareAndSwapInt32(&(p.stat), 0, statError) {
		// TODO: on drop error
		return
	}
	p.actual.OnError(err)
}

func (p *delayElementSubscriber) OnNext(v interface{}) {
	if !atomic.CompareAndSwapInt32(&(p.stat), 0, statComplete) {
		return
	}
	p.v = v
	time.AfterFunc(p.delay, func() {
		p.actual.OnNext(v)
	})
}

func (p *delayElementSubscriber) OnSubscribe(s rs.Subscription) {
	p.s = s
	p.actual.OnSubscribe(p)
	s.Request(rs.RequestInfinite)
}

func (p *delayElementSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&(p.stat), 0, statComplete) {
		p.actual.OnComplete()
	}
}

func newDelayElementSubscriber(actual rs.Subscriber, delay time.Duration, sc scheduler.Scheduler) rs.Subscriber {
	return &delayElementSubscriber{
		delay:  delay,
		actual: actual,
		sc:     sc,
	}
}

type monoDelayElement struct {
	source Mono
	delay  time.Duration
	sc     scheduler.Scheduler
}

func (p *monoDelayElement) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	p.source.SubscribeWith(ctx, newDelayElementSubscriber(s, p.delay, p.sc))
}

func newMonoDelayElement(source Mono, delay time.Duration, sc scheduler.Scheduler) *monoDelayElement {
	return &monoDelayElement{
		source: source,
		delay:  delay,
		sc:     sc,
	}
}
