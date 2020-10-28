package mono

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
)

type processor struct {
	sync.Mutex
	subscriber reactor.Subscriber
	stat       int32
	requested  bool
	executed   int32
	v          internal.Item
}

type processorSubscriber struct {
	parent *processor
	actual reactor.Subscriber
	stat   int32
	s      reactor.Subscription
}

func (p *processor) Success(v Any) {
	p.Lock()
	if p.stat != 0 {
		hooks.Global().OnNextDrop(v)
		p.Unlock()
		return
	}
	p.stat = statComplete
	if p.subscriber == nil || !p.requested {
		p.v = internal.Item{V: v}
		p.Unlock()
		return
	}
	p.Unlock()

	if atomic.AddInt32(&p.executed, 1) != 1 {
		return
	}

	if v != nil {
		p.subscriber.OnNext(v)
	}
	p.subscriber.OnComplete()
}

func (p *processor) Error(e error) {
	p.Lock()
	if p.stat != 0 {
		p.Unlock()
		hooks.Global().OnErrorDrop(e)
		return
	}
	p.stat = statError
	if p.subscriber == nil || !p.requested {
		p.v = internal.Item{E: e}
		p.Unlock()
		return
	}
	p.Unlock()

	if atomic.AddInt32(&p.executed, 1) != 1 {
		return
	}

	p.subscriber.OnError(e)
}

func (p *processor) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	select {
	case <-ctx.Done():
		s.OnError(reactor.ErrSubscribeCancelled)
	default:
		p.Lock()
		if p.subscriber != nil {
			p.Unlock()
			panic("reactor: mono processor can only been subscribed once!")
		}
		p.subscriber = s
		p.Unlock()
		s.OnSubscribe(ctx, &processorSubscriber{
			actual: s,
			parent: p,
		})
	}
}

func (ps *processorSubscriber) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}

	ps.parent.Lock()
	if ps.parent.requested {
		ps.parent.Unlock()
		return
	}
	ps.parent.requested = true
	curStat := ps.parent.stat
	ps.parent.Unlock()

	switch curStat {
	case statError:
		if atomic.AddInt32(&ps.parent.executed, 1) != 1 {
			return
		}
		ps.OnError(ps.parent.v.E)
	case statComplete:
		if atomic.AddInt32(&ps.parent.executed, 1) != 1 {
			return
		}
		v := ps.parent.v.V
		if v != nil {
			ps.OnNext(v)
		}
		ps.OnComplete()
	}
}

func (ps *processorSubscriber) Cancel() {
	atomic.CompareAndSwapInt32(&ps.stat, 0, statCancel)
}

func (ps *processorSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&ps.stat, 0, statComplete) {
		ps.actual.OnComplete()
	}
}

func (ps *processorSubscriber) OnError(e error) {
	if atomic.CompareAndSwapInt32(&ps.stat, 0, statError) {
		ps.actual.OnError(e)
		return
	}
	hooks.Global().OnErrorDrop(e)
}

func (ps *processorSubscriber) OnNext(v Any) {
	if atomic.LoadInt32(&ps.stat) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	ps.actual.OnNext(v)
}

func (ps *processorSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	ps.s = s
	ps.actual.OnSubscribe(ctx, s)
}
