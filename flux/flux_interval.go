package flux

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type intervalSubscription struct {
	actual    rs.Subscriber
	requested *atomic.Int64
	cancelled *atomic.Bool
	done      chan struct{}
	tk        *time.Ticker
	count     *atomic.Int64
}

func (p *intervalSubscription) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
	}
	if p.cancelled.Load() {
		return
	}
	p.requested.Add(int64(n))
}

func (p *intervalSubscription) Cancel() {
	if p.cancelled.CAS(false, true) {
		close(p.done)
	}
}

func (p *intervalSubscription) runOnce() {
	if p.requested.Load() < 1 {
		p.Cancel()
		p.actual.OnError(fmt.Errorf("could not emit tick %d due to lack of requests", p.count))
		return
	}
	current := p.count.Inc()
	p.actual.OnNext(current - 1)
	if p.requested.Load() >= rs.RequestInfinite {
		return
	}
	p.requested.Dec()
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

type fluxInterval struct {
	delay  time.Duration
	period time.Duration
	sc     scheduler.Scheduler
}

func (p *fluxInterval) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	su := &intervalSubscription{
		actual:    s,
		requested: atomic.NewInt64(0),
		cancelled: atomic.NewBool(false),
		done:      make(chan struct{}),
		tk:        time.NewTicker(p.period),
		count:     atomic.NewInt64(0),
	}
	s.OnSubscribe(su)
	p.sc.Worker().Do(func() {
		su.run(ctx)
	})
}

func newFluxInterval(period time.Duration, delay time.Duration, sc scheduler.Scheduler) *fluxInterval {
	if period == 0 {
		panic(fmt.Errorf("invalid interval period: %s", period))
	}
	if sc == nil {
		sc = scheduler.Elastic()
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
