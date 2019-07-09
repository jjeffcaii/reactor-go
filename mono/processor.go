package mono

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

type processor struct {
	subs []rs.Subscriber
	v    interface{}
	e    error
	stat int32
}

func (p *processor) Success(v interface{}) {
	if atomic.CompareAndSwapInt32(&(p.stat), 0, statComplete) {
		p.v = v
		for _, it := range p.subs {
			it.OnNext(v)
			it.OnComplete()
		}
	}
}

func (p *processor) Error(e error) {
	if atomic.CompareAndSwapInt32(&(p.stat), 0, statError) {
		p.e = e
		for _, it := range p.subs {
			it.OnError(e)
		}
	}
}

func (p *processor) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	s := &processorSubscriber{
		actual: actual,
		parent: p,
	}
	actual.OnSubscribe(s)
	p.subs = append(p.subs, s)
}

type processorSubscriber struct {
	parent *processor
	actual rs.Subscriber
	stat   int32
	s      rs.Subscription
	n      int
}

func (p *processorSubscriber) Request(n int) {
	p.n = n
	switch atomic.LoadInt32(&(p.parent.stat)) {
	case statError:
		p.OnError(p.parent.e)
	case statComplete:
		p.OnNext(p.parent.v)
		p.OnComplete()
	}
}

func (p *processorSubscriber) Cancel() {
	atomic.CompareAndSwapInt32(&(p.stat), 0, statCancel)
}

func (p *processorSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&(p.stat), 0, statComplete) {
		p.actual.OnComplete()
	}
}

func (p *processorSubscriber) OnError(e error) {
	if atomic.CompareAndSwapInt32(&(p.stat), 0, statError) {
		p.actual.OnError(e)
	}
}

func (p *processorSubscriber) OnNext(v interface{}) {
	if atomic.LoadInt32(&(p.stat)) != 0 {
		// TODO: discard data
		return
	}
	p.actual.OnNext(v)
}

func (p *processorSubscriber) OnSubscribe(s rs.Subscription) {
	p.s = s
	p.actual.OnSubscribe(s)
}
