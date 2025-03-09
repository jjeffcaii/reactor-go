package flux

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/timerpool"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func Interval(period time.Duration) Flux[int64] {
	return wrap[int64](newFluxInterval(0, period, nil))
}

func Interval2(delay, period time.Duration) Flux[int64] {
	return wrap[int64](newFluxInterval(delay, period, nil))
}

type intervalSubscription struct {
	actual        reactor.Subscriber[int64]
	requested     int64
	cancelled     bool
	done          chan struct{}
	once          sync.Once
	count         int64
	delay, period time.Duration
}

func newIntervalSubscription(actual reactor.Subscriber[int64], delay, period time.Duration) *intervalSubscription {
	return &intervalSubscription{
		actual: actual,
		done:   make(chan struct{}),
		delay:  delay,
		period: period,
	}
}

func (isu *intervalSubscription) Request(ctx context.Context, n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	if isu.cancelled {
		return
	}
	atomic.AddInt64(&isu.requested, int64(n))
}

func (isu *intervalSubscription) Cancel(ctx context.Context) {
	isu.once.Do(func() {
		isu.cancelled = true
		close(isu.done)
	})
}

func (isu *intervalSubscription) runOnce(ctx context.Context) {
	if atomic.LoadInt64(&isu.requested) < 1 {
		isu.Cancel(ctx)
		isu.actual.OnError(ctx, fmt.Errorf("could not emit tick %d due to lack of requests", isu.count))
		return
	}
	current := atomic.AddInt64(&isu.count, 1)
	isu.actual.OnNext(ctx, current-1)
	if atomic.LoadInt64(&isu.requested) >= reactor.RequestInfinite {
		return
	}
	atomic.AddInt64(&isu.requested, -1)
}

func (isu *intervalSubscription) run(ctx context.Context) {
	exe := func(d time.Duration) bool {
		t := timerpool.Get(d)
		defer timerpool.Put(t)

		select {
		case <-ctx.Done():
			isu.actual.OnError(ctx, reactor.NewContextError(ctx.Err()))
			return true
		case <-isu.done:
			return true
		case <-t.C:
			isu.runOnce(ctx)
			return false
		}
	}

	if exe(isu.delay) {
		return
	}

	for {
		if exe(isu.period) {
			break
		}
	}
}

type fluxInterval struct {
	sc            scheduler.Scheduler
	delay, period time.Duration
}

func newFluxInterval(delay, period time.Duration, sc scheduler.Scheduler) *fluxInterval {
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

func (p *fluxInterval) SubscribeWith(ctx context.Context, s reactor.Subscriber[int64]) {
	su := newIntervalSubscription(s, p.delay, p.period)
	s.OnSubscribe(ctx, su)
	err := p.sc.Worker().Do(func() {
		su.run(ctx)
	})
	if err != nil {
		panic(err)
	}
}
