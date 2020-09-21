package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
)

const _delayValue = int64(0)

type delaySubscription struct {
	actual    reactor.Subscriber
	timer     *time.Timer
	done      chan struct{}
	requested chan struct{}
}

func (d *delaySubscription) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	select {
	case <-d.requested:
	default:
		close(d.requested)
	}
}

func (d *delaySubscription) Cancel() {
	select {
	case <-d.done:
	default:
		close(d.done)
		d.timer.Stop()
	}
}

type monoDelay struct {
	delay time.Duration
}

func (p *monoDelay) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	timer := time.NewTimer(p.delay)
	done := make(chan struct{})
	requested := make(chan struct{})

	s := &delaySubscription{
		actual:    actual,
		timer:     timer,
		done:      done,
		requested: requested,
	}

	actual.OnSubscribe(ctx, s)
	go func() {
		defer timer.Stop()
		select {
		case <-ctx.Done():
			close(done)
			actual.OnError(reactor.ErrSubscribeCancelled)
		case <-timer.C:
			close(done)
			// await requested
			<-requested
			// emit value
			actual.OnNext(_delayValue)
			actual.OnComplete()
		case <-done:
			actual.OnError(reactor.ErrSubscribeCancelled)
		}
	}()
}

func newMonoDelay(delay time.Duration) *monoDelay {
	return &monoDelay{
		delay: delay,
	}
}
