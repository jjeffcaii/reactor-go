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

func (is *intervalSubscription) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	if is.cancelled {
		return
	}
	atomic.AddInt64(&is.requested, int64(n))
}

func (is *intervalSubscription) Cancel() {
	is.once.Do(func() {
		is.cancelled = true
		close(is.done)
	})
}

func (is *intervalSubscription) runOnce() {
	if atomic.LoadInt64(&is.requested) < 1 {
		is.Cancel()
		is.actual.OnError(fmt.Errorf("could not emit tick %d due to lack of requests", is.count))
		return
	}
	current := atomic.AddInt64(&is.count, 1)
	is.actual.OnNext(current - 1)
	if atomic.LoadInt64(&is.requested) >= reactor.RequestInfinite {
		return
	}
	atomic.AddInt64(&is.requested, -1)
}

func (is *intervalSubscription) run(ctx context.Context) {
	defer is.tk.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				is.actual.OnError(err)
			}
			return
		case <-is.done:
			return
		case <-is.tk.C:
			is.runOnce()
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
	err := p.sc.Worker().Do(func() {
		su.run(ctx)
	})
	if err != nil {
		panic(err)
	}
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
