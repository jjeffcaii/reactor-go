package mono

import (
	"context"
	"math"
	"time"

	uatomic "go.uber.org/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/timerpool"
)

type monoTimeout[T any] struct {
	source  reactor.RawPublisher[T]
	timeout time.Duration
}

func newMonoTimeout[T any](source reactor.RawPublisher[T], timeout time.Duration) *monoTimeout[T] {
	return &monoTimeout[T]{
		source:  source,
		timeout: timeout,
	}
}

func (m *monoTimeout[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	m.source.SubscribeWith(ctx, &timeoutSubscriber[T]{
		actual:  s,
		timeout: m.timeout,
		done:    make(chan struct{}),
	})
}

type timeoutSubscriber[T any] struct {
	actual  reactor.Subscriber[T]
	timeout time.Duration
	done    chan struct{}
	closed  uatomic.Int32
}

func (ts *timeoutSubscriber[T]) OnComplete(ctx context.Context) {
	if ts.closed.CompareAndSwap(0, math.MaxInt32) || ts.closed.CompareAndSwap(1, math.MaxInt32) {
		close(ts.done)
		ts.actual.OnComplete(ctx)
		return
	}
}

func (ts *timeoutSubscriber[T]) OnError(ctx context.Context, err error) {
	if ts.closed.CompareAndSwap(0, -1) {
		close(ts.done)
		ts.actual.OnError(ctx, err)
		return
	}

	// item is emitted before error reach, should be processed as completed.
	if ts.closed.CompareAndSwap(1, -1) {
		close(ts.done)
		ts.actual.OnComplete(ctx)
		return
	}
	// TODO: ignore error
	//hooks.Global().OnErrorDrop(err)
}

func (ts *timeoutSubscriber[T]) OnNext(ctx context.Context, item T) {
	if !ts.closed.CompareAndSwap(0, 1) {
		// TODO: drop next
		//hooks.Global().OnNextDrop(item)
		return
	}

	ts.actual.OnNext(ctx, item)
}

func (ts *timeoutSubscriber[T]) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	go func(ctx context.Context, timer *time.Timer) {
		defer timerpool.Put(timer)
		select {
		case <-timer.C:
			ts.OnError(ctx, context.DeadlineExceeded)
		case <-ts.done:
		}
	}(ctx, timerpool.Get(ts.timeout))
	ts.actual.OnSubscribe(ctx, su)
}
