package subscribers

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

var (
	_ reactor.Subscriber[any] = (*SwitchValueIfErrorSubscriber[any])(nil)
	_ reactor.Subscription    = (*SwitchValueIfErrorSubscriber[any])(nil)
)

type SwitchValueIfErrorSubscriber[T any] struct {
	actual     reactor.Subscriber[T]
	su         reactor.Subscription
	v          T
	done       int32
	errorCalls int32
}

func NewSwitchValueIfErrorSubscriber[T any](actual reactor.Subscriber[T], value T) *SwitchValueIfErrorSubscriber[T] {
	return &SwitchValueIfErrorSubscriber[T]{
		actual: actual,
		v:      value,
	}
}

func (s *SwitchValueIfErrorSubscriber[T]) Request(ctx context.Context, n int) {
	if atomic.LoadInt32(&s.done) == 0 && s.su != nil {
		s.su.Request(ctx, n)
	}
}

func (s *SwitchValueIfErrorSubscriber[T]) Cancel(ctx context.Context) {
	s.su.Cancel(ctx)
}

func (s *SwitchValueIfErrorSubscriber[T]) OnError(ctx context.Context, err error) {
	if atomic.AddInt32(&s.errorCalls, 1) == 1 {
		// TODO: hooks
		//hooks.Global().OnErrorDrop(err)
		s.actual.OnNext(ctx, s.v)
		s.actual.OnComplete(ctx)
	} else {
		s.actual.OnError(ctx, err)
	}
}

func (s *SwitchValueIfErrorSubscriber[T]) OnNext(ctx context.Context, v T) {
	s.actual.OnNext(ctx, v)
}

func (s *SwitchValueIfErrorSubscriber[T]) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	s.su = su
	s.actual.OnSubscribe(ctx, s)
}

func (s *SwitchValueIfErrorSubscriber[T]) OnComplete(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&s.done, 0, 1) {
		s.actual.OnComplete(ctx)
	}
}
