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
	subs []reactor.Subscriber
	stat int32
	v    Any
}

type processorSubscriber struct {
	parent    *processor
	actual    reactor.Subscriber
	stat      int32
	s         reactor.Subscription
	requested int32
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

func (p *processor) getStat() (stat int32) {
	return atomic.LoadInt32(&p.stat)
}

func (p *processor) Success(v Any) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statComplete) {
		hooks.Global().OnNextDrop(v)
		return
	}

	p.setValue(v)

	p.RLock()
	defer p.RUnlock()
	for i, l := 0, len(p.subs); i < l; i++ {
		s := p.subs[i]
		if v != nil {
			s.OnNext(v)
		}
		s.OnComplete()
	}
}

func (p *processor) Error(e error) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		hooks.Global().OnErrorDrop(e)
		return
	}
	p.setValue(e)
	p.RLock()
	defer p.RUnlock()
	for i, l := 0, len(p.subs); i < l; i++ {
		p.subs[i].OnError(e)
	}
}

func (p *processor) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	select {
	case <-ctx.Done():
		s.OnError(reactor.ErrSubscribeCancelled)
	default:
		s.OnSubscribe(ctx, &processorSubscriber{
			actual: s,
			parent: p,
		})
		p.Lock()
		p.subs = append(p.subs, s)
		p.Unlock()
	}
}

func (p *processorSubscriber) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	if atomic.AddInt32(&p.requested, 1) != 1 {
		return
	}
	v := p.parent.getValue()
	switch p.parent.getStat() {
	case statError:
		p.OnError(v.(error))
	case statComplete:
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
