package mono

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/jjeffcaii/reactor-go"
)

type filterSubscriber[T any] struct {
	ctx       context.Context
	actual    reactor.Subscriber[T]
	predicate reactor.Predicate[T]
	stat      int32
}

func newFilterSubscriber[T any](actual reactor.Subscriber[T], predicate reactor.Predicate[T]) *filterSubscriber[T] {
	return &filterSubscriber[T]{
		actual:    actual,
		predicate: predicate,
	}
}

func (f *filterSubscriber[T]) OnComplete(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&f.stat, 0, statComplete) {
		f.actual.OnComplete(ctx)
	}
}

func (f *filterSubscriber[T]) OnError(ctx context.Context, err error) {
	if !atomic.CompareAndSwapInt32(&f.stat, 0, statError) {
		//TODO: hooks
		//hooks.Global().OnErrorDrop(err)
		return
	}
	f.actual.OnError(ctx, err)
}

func (f *filterSubscriber[T]) OnNext(ctx context.Context, v T) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			f.OnError(ctx, errors.WithStack(e))
		} else {
			f.OnError(ctx, errors.Errorf("%v", rec))
		}
	}()
	if f.predicate(ctx, v) {
		f.actual.OnNext(ctx, v)
		return
	}
	//TODO: discard
	//internal.TryDiscard(f.ctx, v)
}

func (f *filterSubscriber[T]) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	f.ctx = ctx
	f.actual.OnSubscribe(ctx, s)
}

type monoFilter[T any] struct {
	origin    reactor.RawPublisher[T]
	predicate reactor.Predicate[T]
}

func newMonoFilter[T any](origin reactor.RawPublisher[T], predicate reactor.Predicate[T]) monoFilter[T] {
	return monoFilter[T]{
		origin:    origin,
		predicate: predicate,
	}
}

func (m monoFilter[T]) SubscribeWith(ctx context.Context, sub reactor.Subscriber[T]) {
	m.origin.SubscribeWith(ctx, newFilterSubscriber(sub, m.predicate))
}
