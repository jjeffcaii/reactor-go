package mono

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

var (
	_ reactor.RawPublisher[struct{}] = (*processor[struct{}])(nil)
	_ Sink[struct{}]                 = (*processor[struct{}])(nil)
	_ reactor.Disposable             = (*processor[struct{}])(nil)
)

type ProcessorFinallyHook func(reactor.SignalType, reactor.Disposable)

type ProcessorOption[T any] func(*processor[T])

func WithProcessorScheduler[T any](sc scheduler.Scheduler) ProcessorOption[T] {
	return func(p *processor[T]) {
		p.sc = sc
	}
}

func WithProcessorFinallyHook[T any](hook ProcessorFinallyHook) ProcessorOption[T] {
	return func(p *processor[T]) {
		p.hookOnFinally = hook
	}
}

func NewProcessor[T any](opts ...ProcessorOption[T]) (Mono[T], Sink[T], Disposable) {
	p := &processor[T]{}
	for i := range opts {
		if opts[i] == nil {
			continue
		}
		opts[i](p)
	}
	p.doneNotify.L = p.mu.RLocker()
	return wrap[T](p), p, p
}

type processor[T any] struct {
	sc            scheduler.Scheduler
	mu            sync.RWMutex
	item          reactor.Item[T]
	doneNotify    sync.Cond
	hookOnFinally ProcessorFinallyHook
}

func (p *processor[T]) Dispose() {
	//globalProcessorPool.put(p)
}

func (p *processor[T]) Success(item T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.item.Valid {
		// TODO: drop item
		//hooks.Global().OnNextDrop(item)
		return
	}
	p.item.V, p.item.Valid = item, true
	p.doneNotify.Broadcast()
}

func (p *processor[T]) Error(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.item.Valid {
		// TODO: drop error
		//hooks.Global().OnErrorDrop(err)
		return
	}
	p.item.E, p.item.Valid = err, true
	p.doneNotify.Broadcast()
}

func (p *processor[T]) SubscribeWith(ctx context.Context, sub reactor.Subscriber[T]) {
	var s processorSubscriber[T]
	s.source = p
	s.actual = sub

	if p.sc == nil {
		s.OnSubscribe(ctx, &s)
		return
	}

	if err := p.sc.Worker().Do(func() {
		s.OnSubscribe(ctx, &s)
	}); err != nil {
		s.Dispose()
		p.Dispose()
		panic(err)
	}
}

var (
	_ reactor.Subscriber[struct{}] = (*processorSubscriber[struct{}])(nil)
	_ reactor.Subscription         = (*processorSubscriber[struct{}])(nil)
)

type processorSubscriber[T any] struct {
	source    *processor[T]
	actual    reactor.Subscriber[T]
	requested int32
	cancelled int32
}

func (p *processorSubscriber[T]) Request(ctx context.Context, n int) {
	if n <= 0 {
		return
	}
	if !atomic.CompareAndSwapInt32(&p.requested, 0, 1) {
		return
	}

	// return if the subscriber is cancelled
	if atomic.LoadInt32(&p.cancelled) == 1 {
		p.OnError(ctx, reactor.ErrSubscribeCancelled)
		return
	}

	// fetch the item from source publisher
	p.source.mu.RLock()
	item := p.source.item
	if !item.Valid {
		// block for done/cancel if item is not ready
		p.source.doneNotify.Wait()
		item = p.source.item
	}
	p.source.mu.RUnlock()

	// return if the subscriber is cancelled
	if atomic.LoadInt32(&p.cancelled) == 1 {
		p.OnError(ctx, context.Canceled)
		return
	}

	// handle the item
	if !item.Valid {
		p.OnComplete(ctx)
		return
	}

	if item.E != nil {
		p.OnError(ctx, item.E)
		return
	}

	p.OnNext(ctx, item.V)
	p.OnComplete(ctx)
}

func (p *processorSubscriber[T]) Cancel(ctx context.Context) {
	atomic.CompareAndSwapInt32(&p.cancelled, 0, 1)
}

func (p *processorSubscriber[T]) OnComplete(ctx context.Context) {
	p.actual.OnComplete(ctx)
	p.handleFinally(nil)
	p.Dispose()
}

func (p *processorSubscriber[T]) OnError(ctx context.Context, err error) {
	p.actual.OnError(ctx, err)
	p.handleFinally(err)
	p.Dispose()
}

func (p *processorSubscriber[T]) handleFinally(err error) {
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

func (p *processorSubscriber[T]) OnNext(ctx context.Context, item T) {
	p.actual.OnNext(ctx, item)
}

func (p *processorSubscriber[T]) Dispose() {
	//globalProcessorSubscriberPool.put(p)
}

func (p *processorSubscriber[T]) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	select {
	case <-ctx.Done():
		p.OnError(ctx, ctx.Err())
	default:
		p.actual.OnSubscribe(ctx, su)
	}
}
