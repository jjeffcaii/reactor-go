package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"

	uatomic "go.uber.org/atomic"
)

var _ Sink[any] = (*sink[any])(nil)

// Create creates a Mono from a function.
func Create[T any](creation func(ctx context.Context, sink Sink[T])) Mono[T] {
	return wrap[T](newMonoCreate[T](creation))
}

// Sink represents a sink.
type Sink[T any] interface {
	// Success emits an item.
	Success(value T)
	// Error emits an error.
	Error(err error)
}

type monoCreate[T any] func(context.Context, Sink[T])

func newMonoCreate[T any](creation func(context.Context, Sink[T])) monoCreate[T] {
	return func(ctx context.Context, s Sink[T]) {
		select {
		case <-ctx.Done():
			s.Error(ctx.Err())
		default:
			creation(ctx, s)
		}
	}
}

func (mc monoCreate[T]) SubscribeWith(ctx context.Context, sub reactor.Subscriber[T]) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	su := &sink[T]{
		actual: sub,
		cancel: cancel,
	}
	sub.OnSubscribe(ctx, su)
	mc(ctx, su)
}

type sink[T any] struct {
	actual    reactor.Subscriber[T]
	ctx       context.Context
	cancel    context.CancelFunc
	completed uatomic.Bool
}

func (s *sink[T]) Success(v T) {
	if !s.completed.CompareAndSwap(false, true) {
		// TODO: hooks
		//hooks.Global().OnNextDrop(v)
		return
	}
	s.actual.OnNext(s.ctx, v)
	s.actual.OnComplete(s.ctx)
}

func (s *sink[T]) Error(err error) {
	if !s.completed.CompareAndSwap(false, true) {
		// TODO: hooks
		//hooks.Global().OnErrorDrop(err)
		return
	}

	if err == nil {
		s.actual.OnComplete(s.ctx)
		return
	}

	s.actual.OnError(s.ctx, err)
}

func (s *sink[T]) Request(_ context.Context, _ int) {
	// ignore
}

func (s *sink[T]) Cancel(ctx context.Context) {
	if !s.completed.CompareAndSwap(false, true) {
		return
	}

	s.cancel()
	s.actual.OnError(ctx, context.Canceled)
}
