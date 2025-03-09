package flux

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

func Just[T any](values ...T) Flux[T] {
	return wrap[T](newSliceFlux[T](values))
}

type sliceSubscription[T any] struct {
	mu     sync.Mutex
	actual reactor.Subscriber[T]
	values []T
	cursor int32
	flags  uint8
}

func newSliceSubscription[T any](s reactor.Subscriber[T], values []T) *sliceSubscription[T] {
	return &sliceSubscription[T]{
		actual: s,
		values: values,
	}
}

func (ss *sliceSubscription[T]) Request(ctx context.Context, n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	ss.mu.Lock()
	if ss.flags&flagFast != 0 {
		ss.mu.Unlock()
		return
	}
	if n < reactor.RequestInfinite {
		ss.mu.Unlock()
		ss.slowPath(ctx, n)
		return
	}
	ss.flags |= flagFast
	ss.mu.Unlock()
	ss.fastPath(ctx)
}

func (ss *sliceSubscription[T]) Cancel(ctx context.Context) {
	ss.mu.Lock()
	ss.flags |= flagCancel
	ss.mu.Unlock()
}

func (ss *sliceSubscription[T]) isCancelled() (cancelled bool) {
	ss.mu.Lock()
	cancelled = ss.flags&flagCancel != 0
	ss.mu.Unlock()
	return
}

func (ss *sliceSubscription[T]) slowPath(ctx context.Context, n int) {
	for n > 0 {
		next := int(atomic.AddInt32(&ss.cursor, 1))
		if next > len(ss.values) {
			return
		}
		v := ss.values[next-1]
		if next == len(ss.values) {
			ss.actual.OnNext(ctx, v)
			ss.actual.OnComplete(ctx)
			return
		}
		n--
		ss.actual.OnNext(ctx, v)
	}
}

func (ss *sliceSubscription[T]) fastPath(ctx context.Context) {
	i, l := int(ss.cursor), len(ss.values)
	if i >= l {
		return
	}
	for ; i < l; i++ {
		if ss.isCancelled() {
			return
		}
		ss.actual.OnNext(ctx, ss.values[i])
	}
	if ss.isCancelled() {
		return
	}
	ss.actual.OnComplete(ctx)
}

type fluxSlice[T any] []T

func newSliceFlux[T any](values []T) fluxSlice[T] {
	return values
}

func (p fluxSlice[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	if len(p) < 1 {
		s.OnComplete(ctx)
		return
	}
	su := newSliceSubscription[T](s, p)
	s.OnSubscribe(ctx, su)
}
