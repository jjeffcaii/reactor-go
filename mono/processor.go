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
	v      interface{}
	e      error
	done   bool
	subs   *sync.Map
	locker sync.RWMutex
}

func (p *processor) stat() (stat int32) {
	p.locker.RLock()
	if p.e != nil {
		stat = statError
	} else if p.done {
		stat = statComplete
	}
	p.locker.RUnlock()
	return
}

func (p *processor) Success(v interface{}) {
	p.locker.Lock()
	if p.done {
		p.locker.Unlock()
		hooks.Global().OnNextDrop(v)
		return
	}
	p.done = true
	p.v = v
	p.locker.Unlock()
	p.subs.Range(func(key, value interface{}) bool {
		s := key.(rs.Subscriber)
		if v != nil {
			s.OnNext(v)
		}
		s.OnComplete()
		return true
	})
}

func (p *processor) Error(e error) {
	p.locker.Lock()
	if p.done {
		p.locker.Unlock()
		hooks.Global().OnErrorDrop(e)
		return
	}
	p.done = true
	p.e = e
	p.locker.Unlock()
	p.subs.Range(func(key, value interface{}) bool {
		s := key.(rs.Subscriber)
		s.OnError(e)
		return true
	})
}

func (p *processor) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	s := &processorSubscriber{
		actual: actual,
		parent: p,
	}
	actual = internal.NewCoreSubscriber(ctx, s)
	actual.OnSubscribe(s)
	p.subs.Store(actual, true)
}

type processorSubscriber struct {
	parent    *processor
	actual    rs.Subscriber
	stat      int32
	s         rs.Subscription
	requested int32
}

func (p *processorSubscriber) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
	}
	if atomic.AddInt32(&(p.requested), 1) != 1 {
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
		return
	}
	hooks.Global().OnErrorDrop(e)
}

func (p *processorSubscriber) OnNext(v interface{}) {
	if atomic.LoadInt32(&(p.stat)) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	p.actual.OnNext(v)
}

func (p *processorSubscriber) OnSubscribe(s rs.Subscription) {
	p.s = s
	p.actual.OnSubscribe(s)
}
