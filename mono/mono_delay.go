package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type delaySubscriber struct {
	actual    reactor.Subscriber
	requested bool
}

func (p *delaySubscriber) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	p.requested = true
}

func (*delaySubscriber) Cancel() {
	panic("implement me")
}

type monoDelay struct {
	delay time.Duration
	sc    scheduler.Scheduler
}

func (p *monoDelay) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	s := &delaySubscriber{
		actual: actual,
	}
	actual.OnSubscribe(s)

	time.AfterFunc(p.delay, func() {
		if p.sc == nil {
			actual.OnNext(int64(0))
			actual.OnComplete()
			return
		}
		p.sc.Worker().Do(func() {
			actual.OnNext(int64(0))
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
