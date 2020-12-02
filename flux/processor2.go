package flux

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal/buffer"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

var (
	_ reactor.Disposable   = (*processor)(nil)
	_ reactor.RawPublisher = (*processor)(nil)
	_ Sink                 = (*processor)(nil)

	_ reactor.Subscription = (*processorSubscription)(nil)
	_ reactor.Subscriber   = (*processorSubscription)(nil)
)

type pstat int32

const (
	_ pstat = iota
	pstatComplete
	pstatError
	pstatCancelling
	pstatCancelled
)

type processorSubscription struct {
	actual    reactor.Subscriber
	b         *buffer.Unbounded
	state     int32
	requested int64
	draining  int32
	canceller chan struct{}
}

func (x *processorSubscription) OnComplete() {
	if atomic.CompareAndSwapInt32(&x.state, 0, int32(pstatComplete)) {
		x.actual.OnComplete()
	}
}

func (x *processorSubscription) OnError(err error) {
	if !atomic.CompareAndSwapInt32(&x.state, 0, int32(pstatError)) {
		hooks.Global().OnErrorDrop(err)
		return
	}
	x.actual.OnError(err)
}

func (x *processorSubscription) OnNext(any reactor.Any) {
	if atomic.LoadInt32(&x.state) != 0 {
		hooks.Global().OnNextDrop(any)
		return
	}
	x.actual.OnNext(any)
}

func (x *processorSubscription) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	x.actual.OnSubscribe(ctx, su)
}

func (x *processorSubscription) Request(n int) {
	if n < 1 {
		return
	}
	if atomic.AddInt64(&x.requested, int64(n)) > 0 {
		x.drain()
	}
}

func (x *processorSubscription) Cancel() {
	if !atomic.CompareAndSwapInt32(&x.state, 0, int32(pstatCancelling)) {
		return
	}
	close(x.canceller)
	x.b.Dispose()
	x.clear()
}

func (x *processorSubscription) drain() {
	if !atomic.CompareAndSwapInt32(&x.draining, 0, 1) {
		return
	}
	defer atomic.CompareAndSwapInt32(&x.draining, 1, 0)

	for atomic.AddInt64(&x.requested, -1) >= 0 {
		select {
		case next, ok := <-x.b.Get():
			x.b.Load()
			if !ok {
				x.OnComplete()
				return
			}
			item := next.(reactor.Item)
			if item.E != nil {
				x.OnError(item.E)
			} else if item.V != nil {
				x.OnNext(item.V)
			}
		case <-x.canceller:
			x.clear()
			return
		}
	}
}

func (x *processorSubscription) clear() {
	if !atomic.CompareAndSwapInt32(&x.state, int32(pstatCancelling), int32(pstatCancelled)) {
		return
	}
	// drop items in buffer
	for next := range x.b.Get() {
		x.b.Load()
		it := next.(reactor.Item)
		if it.E != nil {
			hooks.Global().OnErrorDrop(it.E)
		} else if it.V != nil {
			hooks.Global().OnNextDrop(it.V)
		}
	}
	// emit error signal
	x.actual.OnError(reactor.ErrSubscribeCancelled)
}

type processor struct {
	sc         scheduler.Scheduler
	b          *buffer.Unbounded
	subscribed uint32
}

func (xp *processor) Dispose() {
	xp.b.Dispose()
}

func (xp *processor) Complete() {
	xp.b.Dispose()
}

func (xp *processor) Error(err error) {
	ok := xp.b.Put(reactor.Item{E: err})
	if !ok {
		hooks.Global().OnErrorDrop(err)
		return
	}
	xp.b.Dispose()
}

func (xp *processor) Next(any Any) {
	if !xp.b.Put(reactor.Item{V: any}) {
		hooks.Global().OnNextDrop(any)
	}
}

func (xp *processor) SubscribeWith(ctx context.Context, sub reactor.Subscriber) {
	if atomic.AddUint32(&xp.subscribed, 1) > 1 {
		handler := func() {
			sub.OnError(errors.New("process has been subscribed"))
		}
		if err := xp.sc.Worker().Do(handler); err != nil {
			handler()
		}
		return
	}
	su := &processorSubscription{
		actual:    sub,
		b:         xp.b,
		canceller: make(chan struct{}),
	}
	handler := func() {
		su.OnSubscribe(ctx, su)
	}
	if err := xp.sc.Worker().Do(handler); err != nil {
		handler()
	}
}

func newProcessor(sc scheduler.Scheduler) *processor {
	return &processor{
		b:  buffer.NewUnbounded(),
		sc: sc,
	}
}
