package mono

import (
	"context"
	"math"
	"sync/atomic"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

type monoTimeout struct {
	source  reactor.RawPublisher
	timeout time.Duration
}

func newMonoTimeout(source reactor.RawPublisher, timeout time.Duration) *monoTimeout {
	return &monoTimeout{
		source:  source,
		timeout: timeout,
	}
}

func (m *monoTimeout) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	m.source.SubscribeWith(ctx, &timeoutSubscriber{
		actual:  s,
		timeout: m.timeout,
		done:    make(chan struct{}),
	})
}

type timeoutSubscriber struct {
	actual  reactor.Subscriber
	timeout time.Duration
	done    chan struct{}
	closed  int32
}

func (t *timeoutSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&t.closed, 0, math.MaxInt32) || atomic.CompareAndSwapInt32(&t.closed, 1, math.MaxInt32) {
		close(t.done)
		t.actual.OnComplete()
	}
}

func (t *timeoutSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&t.closed, 0, -1) {
		close(t.done)
		t.actual.OnError(err)
	} else if atomic.CompareAndSwapInt32(&t.closed, 1, -1) {
		close(t.done)
	} else {
		hooks.Global().OnErrorDrop(err)
	}
}

func (t *timeoutSubscriber) OnNext(any reactor.Any) {
	if atomic.CompareAndSwapInt32(&t.closed, 0, 1) {
		t.actual.OnNext(any)
	} else {
		hooks.Global().OnNextDrop(any)
	}
}

func (t *timeoutSubscriber) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	timer := time.NewTimer(t.timeout)
	go func() {
		defer timer.Stop()
		select {
		case <-timer.C:
			t.OnError(reactor.ErrSubscribeCancelled)
		case <-t.done:
		}
	}()
	t.actual.OnSubscribe(ctx, subscription)
}
