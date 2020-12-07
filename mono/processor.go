package mono

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type ProcessorFinallyHook func(reactor.SignalType, reactor.Disposable)

var (
	_ reactor.RawPublisher = (*processor)(nil)
	_ Sink                 = (*processor)(nil)
	_ reactor.Disposable   = (*processor)(nil)
)

var (
	globalProcessorPool           processorPool
	globalProcessorSubscriberPool processorSubscriberPool
)

type processorPool struct {
	inner sync.Pool
}

func (pp *processorPool) get() *processor {
	if exist, _ := pp.inner.Get().(*processor); exist != nil {
		return exist
	}
	p := &processor{}
	p.doneNotify.L = p.mu.RLocker()
	return p
}

func (pp *processorPool) put(p *processor) {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.item = nil
	p.hookOnFinally = nil
	p.mu.Unlock()
	pp.inner.Put(p)
}

type processorSubscriberPool struct {
	inner sync.Pool
}

func (pp *processorSubscriberPool) get() *processorSubscriber {
	if exist, _ := pp.inner.Get().(*processorSubscriber); exist != nil {
		return exist
	}
	return &processorSubscriber{}
}

func (pp *processorSubscriberPool) put(p *processorSubscriber) {
	if p == nil {
		return
	}
	p.actual = nil
	p.source = nil
	atomic.StoreInt32(&p.requested, 0)
	pp.inner.Put(p)
}

type processor struct {
	sc            scheduler.Scheduler
	mu            sync.RWMutex
	item          *reactor.Item
	doneNotify    sync.Cond
	hookOnFinally ProcessorFinallyHook
}

func (p *processor) Dispose() {
	globalProcessorPool.put(p)
}

func (p *processor) Success(any Any) {
	p.mu.Lock()
	if p.item != nil {
		p.mu.Unlock()
		hooks.Global().OnNextDrop(any)
		return
	}
	p.item = &reactor.Item{V: any}
	p.doneNotify.Broadcast()
	p.mu.Unlock()
}

func (p *processor) Error(err error) {
	p.mu.Lock()
	if p.item != nil {
		p.mu.Unlock()
		hooks.Global().OnErrorDrop(err)
		return
	}
	p.item = &reactor.Item{E: err}
	p.doneNotify.Broadcast()
	p.mu.Unlock()
}

func (p *processor) SubscribeWith(ctx context.Context, sub reactor.Subscriber) {
	s := globalProcessorSubscriberPool.get()
	s.source = p
	s.actual = sub
	atomic.StoreInt32(&s.cancelled, 0)

	if p.sc == nil {
		s.OnSubscribe(ctx, s)
		return
	}

	if err := p.sc.Worker().Do(func() {
		s.OnSubscribe(ctx, s)
	}); err != nil {
		s.Dispose()
		p.Dispose()
		panic(err)
	}
}

var (
	_ reactor.Subscriber   = (*processorSubscriber)(nil)
	_ reactor.Subscription = (*processorSubscriber)(nil)
)

type processorSubscriber struct {
	source    *processor
	actual    reactor.Subscriber
	requested int32
	cancelled int32
}

func (p *processorSubscriber) Request(n int) {
	if n <= 0 {
		return
	}
	if !atomic.CompareAndSwapInt32(&p.requested, 0, 1) {
		return
	}

	// return if the subscriber is cancelled
	if atomic.LoadInt32(&p.cancelled) == 1 {
		p.OnError(reactor.ErrSubscribeCancelled)
		return
	}

	var item *reactor.Item

	// fetch the item from source publisher
	p.source.mu.RLock()
	item = p.source.item
	if item == nil {
		// block for done/cancel if item is not ready
		p.source.doneNotify.Wait()
		item = p.source.item
	}
	p.source.mu.RUnlock()

	// return if the subscriber is cancelled
	if atomic.LoadInt32(&p.cancelled) == 1 {
		p.OnError(reactor.ErrSubscribeCancelled)
		return
	}

	// handle the item
	if item == nil {
		p.OnComplete()
	} else if item.E != nil {
		p.OnError(item.E)
	} else if item.V != nil {
		p.OnNext(item.V)
		p.OnComplete()
	} else {
		p.OnComplete()
	}
}

func (p *processorSubscriber) Cancel() {
	atomic.CompareAndSwapInt32(&p.cancelled, 0, 1)
}

func (p *processorSubscriber) OnComplete() {
	p.actual.OnComplete()
	p.handleFinally(nil)
	p.Dispose()
}

func (p *processorSubscriber) OnError(err error) {
	p.actual.OnError(err)
	p.handleFinally(err)
	p.Dispose()
}

func (p *processorSubscriber) handleFinally(err error) {
	p.source.mu.Lock()
	fn := p.source.hookOnFinally
	p.source.mu.Unlock()
	if fn == nil {
		return
	}
	if err == nil {
		fn(reactor.SignalTypeComplete, p.source)
	} else if reactor.IsCancelledError(err) {
		fn(reactor.SignalTypeCancel, p.source)
	} else {
		fn(reactor.SignalTypeError, p.source)
	}
}

func (p *processorSubscriber) OnNext(any reactor.Any) {
	p.actual.OnNext(any)
}

func (p *processorSubscriber) Dispose() {
	globalProcessorSubscriberPool.put(p)
}

func (p *processorSubscriber) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	select {
	case <-ctx.Done():
		p.OnError(reactor.NewContextError(ctx.Err()))
	default:
		p.actual.OnSubscribe(ctx, su)
	}
}
