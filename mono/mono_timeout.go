package mono

import (
	"context"
	"time"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
)

type monoTimeout struct {
	source  reactor.RawPublisher
	timeout time.Duration
}

type timeoutSubscriber struct {
	actual  reactor.Subscriber
	timeout time.Duration
	done    chan struct{}
}

func (t *timeoutSubscriber) OnComplete() {
	select {
	case <-t.done:
	default:
		close(t.done)
		t.actual.OnComplete()
	}
}

func (t *timeoutSubscriber) OnError(err error) {
	select {
	case <-t.done:
		hooks.Global().OnErrorDrop(err)
	default:
		close(t.done)
		t.actual.OnError(err)
	}
}

func (t *timeoutSubscriber) OnNext(any reactor.Any) {
	select {
	case <-t.done:
		hooks.Global().OnNextDrop(any)
	default:
		t.actual.OnNext(any)
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

func (m *monoTimeout) SubscribeWith(ctx context.Context, subscriber reactor.Subscriber) {
	subscriber = internal.ExtractRawSubscriber(subscriber)
	ts := &timeoutSubscriber{
		actual:  subscriber,
		timeout: m.timeout,
		done:    make(chan struct{}),
	}
	subscriber = internal.NewCoreSubscriber(ts)
	m.source.SubscribeWith(ctx, subscriber)
}

func newMonoTimeout(source reactor.RawPublisher, timeout time.Duration) *monoTimeout {
	return &monoTimeout{
		source:  source,
		timeout: timeout,
	}
}
