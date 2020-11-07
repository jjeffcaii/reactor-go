package subscribers

import (
	"context"
	"github.com/jjeffcaii/reactor-go"
	"sync/atomic"
)

type DoCreateIfErrorSubscriber struct {
	actual    reactor.Subscriber
	v  		  reactor.Any
	s         reactor.Subscription
	done      int32
	requested int
}

func NewDoCreateIfErrorSubscriber(actual reactor.Subscriber, value reactor.Any) *DoCreateIfErrorSubscriber {
	return &DoCreateIfErrorSubscriber{
		actual:    actual,
		v: value,
	}
}

func (d *DoCreateIfErrorSubscriber) Request(n int) {
	if atomic.LoadInt32(&d.done) == 0 {
		d.requested = n
		if d.s != nil {
			d.s.Request(n)
		}
	}
}

func (d *DoCreateIfErrorSubscriber) Cancel() {
	d.s.Cancel()
}

func (d *DoCreateIfErrorSubscriber) OnError(_ error) {
	d.OnNext(d.v)
	d.OnComplete()
}

func (d *DoCreateIfErrorSubscriber) OnNext(v reactor.Any) {
	d.actual.OnNext(v)
}

func (d *DoCreateIfErrorSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	d.s = s
	d.actual.OnSubscribe(ctx, d)
}

func (d *DoCreateIfErrorSubscriber) OnComplete() {
	if atomic.AddInt32(&d.done, 1) == 1 {
		d.actual.OnComplete()
	}
}


