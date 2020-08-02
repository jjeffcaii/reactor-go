package flux

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type intervalSubscription struct {
	actual    reactor.Subscriber
	requested int64
	cancelled bool
	done      chan struct{}
	tk        *time.Ticker
	once      sync.Once
	count     int64
}

func (p *intervalSubscription) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	if p.cancelled {
		return
	}
	atomic.AddInt64(&p.requested, int64(n))
}

func (p *intervalSubscription) Cancel() {
	p.once.Do(func() {
		p.cancelled = true
		close(p.done)
	})
}

func (p *intervalSubscription) runOnce() {
	if atomic.LoadInt64(&p.requested) < 1 {
		p.Cancel()
		p.actual.OnError(fmt.Errorf("could not emit tick %d due to lack of requests", p.count))
		return
	}
	current := atomic.AddInt64(&p.count, 1)
	p.actual.OnNext(current - 1)
	if atomic.LoadInt64(&p.requested) >= reactor.RequestInfinite {
		return
	}
	atomic.AddInt64(&p.requested, -1)
}

func (p *intervalSubscription) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				p.actual.OnError(err)
			}
			return
		case <-p.done:
			return
		case <-p.tk.C:
			p.runOnce()
		}
	}
}

func newIntervalSubscription(actual reactor.Subscriber, period time.Duration) *intervalSubscription {
	return &intervalSubscription{
		actual: actual,
		done:   make(chan struct{}),
		tk:     time.NewTicker(period),
	}
}

type fluxInterval struct {
	delay  time.Duration
	period time.Duration
	sc     scheduler.Scheduler
}

func (p *fluxInterval) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	su := newIntervalSubscription(s, p.period)
	s.OnSubscribe(su)
	p.sc.Worker().Do(func() {
		su.run(ctx)
	})
}

func newFluxInterval(period time.Duration, delay time.Duration, sc scheduler.Scheduler) *fluxInterval {
	if period == 0 {
		panic(fmt.Sprintf("invalid interval period: %s", period))
	}
	if sc == nil {
		sc = scheduler.Parallel()
	}
	if delay == 0 {
		delay = period
	}
	return &fluxInterval{
		period: period,
		delay:  delay,
		sc:     sc,
	}
}
