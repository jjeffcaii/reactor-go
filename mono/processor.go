package mono

import (
	"context"
	"sync"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
	"go.uber.org/atomic"
)

type processor struct {
	v      interface{}
	e      error
	done   bool
	subs   *internal.Vector
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
	for _, it := range p.subs.Snapshot() {
		s := it.(rs.Subscriber)
		if v != nil {
			s.OnNext(v)
		}
		s.OnComplete()
	}
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
	for _, it := range p.subs.Snapshot() {
		it.(rs.Subscriber).OnError(e)
	}
}

func (p *processor) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	s := &processorSubscriber{
		actual:    actual,
		parent:    p,
		requested: atomic.NewInt32(0),
		stat:      atomic.NewInt32(0),
	}
	actual = internal.NewCoreSubscriber(ctx, s)
	actual.OnSubscribe(s)
	p.subs.Add(actual)
}

type processorSubscriber struct {
	parent    *processor
	actual    rs.Subscriber
	stat      *atomic.Int32
	s         rs.Subscription
	requested *atomic.Int32
}

func (p *processorSubscriber) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
	}
	if p.requested.Inc() != 1 {
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
	p.stat.CAS(0, statCancel)
}

func (p *processorSubscriber) OnComplete() {
	if p.stat.CAS(0, statComplete) {
		p.actual.OnComplete()
	}
}

func (p *processorSubscriber) OnError(e error) {
	if p.stat.CAS(0, statError) {
		p.actual.OnError(e)
		return
	}
	hooks.Global().OnErrorDrop(e)
}

func (p *processorSubscriber) OnNext(v interface{}) {
	if p.stat.Load() != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	p.actual.OnNext(v)
}

func (p *processorSubscriber) OnSubscribe(s rs.Subscription) {
	p.s = s
	p.actual.OnSubscribe(s)
}

func newProcessor() *processor {
	return &processor{
		subs: internal.NewVector(),
	}
}
