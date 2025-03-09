package subscribers

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

var _ reactor.Subscriber[any] = (*DoFinallySubscriber[any])(nil)

type DoFinallySubscriber[T any] struct {
	actual    reactor.Subscriber[T]
	onFinally reactor.FnOnFinally
	s         reactor.Subscription
	done      int32
}

func NewDoFinallySubscriber[T any](actual reactor.Subscriber[T], onFinally reactor.FnOnFinally) *DoFinallySubscriber[T] {
	var sub DoFinallySubscriber[T]
	sub.actual = actual
	sub.onFinally = onFinally
	return &sub
}

func (d *DoFinallySubscriber[T]) Request(ctx context.Context, n int) {
	if atomic.LoadInt32(&d.done) == 0 {
		d.s.Request(ctx, n)
	}
}

func (d *DoFinallySubscriber[T]) Cancel(ctx context.Context) {
	if d.s == nil {
		return
	}
	d.s.Cancel(ctx)
	d.runFinally(ctx, reactor.SignalTypeCancel)
}

func (d *DoFinallySubscriber[T]) OnError(ctx context.Context, err error) {
	d.actual.OnError(ctx, err)
	if reactor.IsCancelledError(err) {
		d.runFinally(ctx, reactor.SignalTypeCancel)
	} else {
		d.runFinally(ctx, reactor.SignalTypeError)
	}
}

func (d *DoFinallySubscriber[T]) OnNext(ctx context.Context, v T) {
	d.actual.OnNext(ctx, v)
}

func (d *DoFinallySubscriber[T]) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	select {
	case <-ctx.Done():
		d.OnError(ctx, reactor.NewContextError(ctx.Err()))
	default:
		d.s = s
		d.actual.OnSubscribe(ctx, d)
	}
}

func (d *DoFinallySubscriber[T]) OnComplete(ctx context.Context) {
	d.actual.OnComplete(ctx)
	d.runFinally(ctx, reactor.SignalTypeComplete)
}

func (d *DoFinallySubscriber[T]) runFinally(ctx context.Context, sig reactor.SignalType) {
	if atomic.AddInt32(&d.done, 1) == 1 {
		d.onFinally(ctx, sig)
	}
}
