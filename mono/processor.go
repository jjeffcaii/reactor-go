package mono

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

type processor struct {
	sync.RWMutex
	subscriber reactor.Subscriber
	stat       int32
	requested  int32
	v          Any
}

type processorSubscriber struct {
	parent *processor
	actual reactor.Subscriber
	stat   int32
	s      reactor.Subscription
}

func (p *processor) getValue() Any {
	p.RLock()
	defer p.RUnlock()
	return p.v
}

func (p *processor) setValue(v Any) {
	p.Lock()
	defer p.Unlock()
	p.v = v
}

func (p *processor) Success(v Any) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statComplete) {
		hooks.Global().OnNextDrop(v)
		return
	}

	p.RLock()
	if p.subscriber == nil {
		p.RUnlock()
		p.setValue(v)
		return
	}
	p.RUnlock()

	if atomic.LoadInt32(&p.requested) > 0 {
		if v != nil {
			p.subscriber.OnNext(v)
		}
		p.subscriber.OnComplete()
	} else {
		p.setValue(v)
	}
}

func (p *processor) Error(e error) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		hooks.Global().OnErrorDrop(e)
		return
	}

	p.RLock()
	if p.subscriber == nil {
		p.RUnlock()
		p.setValue(e)
		return
	}
	p.RUnlock()

	if atomic.LoadInt32(&p.requested) > 0 {
		p.subscriber.OnError(e)
	} else {
		p.setValue(e)
	}
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

func (p *processor) getStat() (stat int32) {
	return atomic.LoadInt32(&p.stat)
}

func (p *processorSubscriber) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}

	if atomic.AddInt32(&p.parent.requested, 1) != 1 {
		return
	}

	switch p.parent.getStat() {
	case statError:
		v := p.parent.getValue()
		p.OnError(v.(error))
	case statComplete:
		v := p.parent.getValue()
		if v != nil {
			p.OnNext(v)
		}
		p.OnComplete()
	}

}

func (p *processorSubscriber) Cancel() {
	atomic.CompareAndSwapInt32(&p.stat, 0, statCancel)
}

func (p *processorSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&p.stat, 0, statComplete) {
		p.actual.OnComplete()
	}
}

func (p *processorSubscriber) OnError(e error) {
	if atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		p.actual.OnError(e)
		return
	}
	hooks.Global().OnErrorDrop(e)
}

func (p *processorSubscriber) OnNext(v Any) {
	if atomic.LoadInt32(&p.stat) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	p.actual.OnNext(v)
}

func (p *processorSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	p.s = s
	p.actual.OnSubscribe(ctx, s)
}
