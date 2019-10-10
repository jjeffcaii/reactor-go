package mono

import (
	"context"
	"time"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"go.uber.org/atomic"
)

type delayElementSubscriber struct {
	delay  time.Duration
	sc     scheduler.Scheduler
	s      rs.Subscription
	actual rs.Subscriber
	stat   *atomic.Int32
	v      interface{}
}

func (p *delayElementSubscriber) Request(n int) {
	p.s.Request(n)
}

func (p *delayElementSubscriber) Cancel() {
	if p.stat.CAS(0, statCancel) {
		p.s.Cancel()
	}
}

func (p *delayElementSubscriber) OnError(err error) {
	if p.stat.CAS(0, statError) {
		p.actual.OnError(err)
		return
	}
	hooks.Global().OnErrorDrop(err)
}

func (p *delayElementSubscriber) OnNext(v interface{}) {
	if p.stat.Load() != 0 {
		hooks.Global().OnNextDrop(v)
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
	if p.stat.CAS(0, statComplete) {
		p.actual.OnComplete()
	}
}

type monoDelayElement struct {
	source rs.RawPublisher
	delay  time.Duration
	sc     scheduler.Scheduler
}

func (p *monoDelayElement) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	p.source.SubscribeWith(ctx, internal.NewCoreSubscriber(ctx, &delayElementSubscriber{
		delay:  p.delay,
		actual: actual,
		sc:     p.sc,
		stat:   atomic.NewInt32(0),
	}))
}

func newMonoDelayElement(source rs.RawPublisher, delay time.Duration, sc scheduler.Scheduler) *monoDelayElement {
	return &monoDelayElement{
		source: source,
		delay:  delay,
		sc:     sc,
	}
}
