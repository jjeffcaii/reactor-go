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
	sync.RWMutex
	v    Any
	e    error
	done bool
	subs *internal.Vector
}

func (p *processor) stat() (stat int32) {
	p.RLock()
	if p.e != nil {
		stat = statError
	} else if p.done {
		stat = statComplete
	}
	p.RUnlock()
	return
}

func (p *processor) Success(v Any) {
	p.Lock()
	if p.done {
		p.Unlock()
		hooks.Global().OnNextDrop(v)
		return
	}
	p.done = true
	p.v = v
	p.Unlock()
	for _, it := range p.subs.Snapshot() {
		s := it.(reactor.Subscriber)
		if v != nil {
			s.OnNext(v)
		}
		s.OnComplete()
	}
}

func (p *processor) Error(e error) {
	p.Lock()
	if p.done {
		p.Unlock()
		hooks.Global().OnErrorDrop(e)
		return
	}
	p.done = true
	p.e = e
	p.Unlock()
	for _, it := range p.subs.Snapshot() {
		it.(reactor.Subscriber).OnError(e)
	}
}

func (p *processor) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	s := &processorSubscriber{
		actual: actual,
		parent: p,
	}
	actual = internal.NewCoreSubscriber(s)
	actual.OnSubscribe(ctx, s)
	p.subs.Add(actual)
}

type processorSubscriber struct {
	parent    *processor
	actual    reactor.Subscriber
	stat      int32
	s         reactor.Subscription
	requested int32
}

func (p *processorSubscriber) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	if atomic.AddInt32(&p.requested, 1) != 1 {
		return
	}
	switch p.parent.stat() {
	case statError:
		p.OnError(p.parent.e)
	case statComplete:
		if p.parent.v != nil {
			p.OnNext(p.parent.v)
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

func newProcessor() *processor {
	return &processor{
		subs: internal.NewVector(),
	}
}
