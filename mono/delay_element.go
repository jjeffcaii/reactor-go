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

var (
	_ reactor.Subscription = (*delayElementSubscriber)(nil)
	_ reactor.Subscriber   = (*delayElementSubscriber)(nil)
)

const (
	_ = iota
	_delayNextStatRunning
	_delayNextStatConfirmed
	_delayNextStatFinish
)

type delayElementSubscriber struct {
	delay     time.Duration
	sc        scheduler.Scheduler
	ctx       context.Context
	s         reactor.Subscription
	actual    reactor.Subscriber
	done      int32
	nextStat  int32
	cancelled chan struct{}
}

func (d *delayElementSubscriber) Request(n int) {
	d.s.Request(n)
}

func (d *delayElementSubscriber) Cancel() {
	d.s.Cancel()
	internal.SafeCloseDone(d.cancelled)
}

func (d *delayElementSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	d.s = s
	d.ctx = ctx
	d.actual.OnSubscribe(ctx, d)
}

func (d *delayElementSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&d.done, 0, 1) {
		d.actual.OnError(err)
		return
	}
	hooks.Global().OnErrorDrop(err)
}

func (d *delayElementSubscriber) OnNext(v Any) {
	if atomic.LoadInt32(&d.done) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	atomic.StoreInt32(&d.nextStat, _delayNextStatRunning)

	if err := d.sc.Worker().Do(func() {
		select {
		case <-d.ctx.Done():
			d.OnError(reactor.ErrSubscribeCancelled)
		case <-time.After(d.delay):
			d.actual.OnNext(v)
			if atomic.CompareAndSwapInt32(&d.nextStat, _delayNextStatConfirmed, _delayNextStatFinish) {
				d.actual.OnComplete()
			}
		case <-d.cancelled:
			d.OnError(reactor.ErrSubscribeCancelled)
		}
	}); err != nil {
		d.OnError(err)
	}
}

func (d *delayElementSubscriber) OnComplete() {
	if !atomic.CompareAndSwapInt32(&d.done, 0, 1) {
		return
	}
	if !atomic.CompareAndSwapInt32(&d.nextStat, _delayNextStatRunning, _delayNextStatConfirmed) {
		d.actual.OnComplete()
	}
}

func newDelayElementSubscriber(actual reactor.Subscriber, delay time.Duration, sc scheduler.Scheduler) reactor.Subscriber {
	return &delayElementSubscriber{
		delay:     delay,
		actual:    actual,
		sc:        sc,
		cancelled: make(chan struct{}),
	}
}

type monoDelayElement struct {
	source reactor.RawPublisher
	delay  time.Duration
	sc     scheduler.Scheduler
}

func (p *monoDelayElement) Parent() reactor.RawPublisher {
	return p.source
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
