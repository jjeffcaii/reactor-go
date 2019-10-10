package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"go.uber.org/atomic"
)

type delaySubscriber struct {
	actual    rs.Subscriber
	requested *atomic.Bool
	cancelled *atomic.Bool
}

func (p *delaySubscriber) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
	}
	p.requested.CAS(false, true)
}

func (p *delaySubscriber) Cancel() {
	p.cancelled.Store(true)
}

type monoDelay struct {
	delay time.Duration
	sc    scheduler.Scheduler
}

func (p *monoDelay) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	s := &delaySubscriber{
		actual:    actual,
		requested: atomic.NewBool(false),
		cancelled: atomic.NewBool(false),
	}
	actual.OnSubscribe(s)

	time.AfterFunc(p.delay, func() {
		p.sc.Worker().Do(func() {
			if s.cancelled.Load() {
				actual.OnError(rs.ErrSubscribeCancelled)
				return
			}
			if s.requested.Load() {
				actual.OnNext(int64(0))
			}
			actual.OnComplete()
		})
	})
}

func newMonoDelay(delay time.Duration, sc scheduler.Scheduler) *monoDelay {
	return &monoDelay{
		delay: delay,
		sc:    sc,
	}
}
