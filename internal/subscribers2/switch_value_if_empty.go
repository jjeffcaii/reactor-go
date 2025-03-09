package subscribers

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

var (
	_ reactor.Subscriber[any] = (*SwitchValueIfEmptySubscriber[any])(nil)
	_ reactor.Subscription    = (*SwitchValueIfEmptySubscriber[any])(nil)
)

type SwitchValueIfEmptySubscriber[T any] struct {
	ctx         context.Context
	actual      reactor.Subscriber[T]
	alternative T
	nextOnce    bool
	cancelled   bool
	su          reactor.Subscription
	requested   int
}

func NewSwitchValueIfEmptySubscriber[T any](actual reactor.Subscriber[T], alternative T) *SwitchValueIfEmptySubscriber[T] {
	return &SwitchValueIfEmptySubscriber[T]{
		actual:      actual,
		alternative: alternative,
	}
}

func (s *SwitchValueIfEmptySubscriber[T]) Request(ctx context.Context, n int) {
	if n < 1 {
		return
	}
	s.requested = n
	if s.su != nil {
		s.su.Request(ctx, n)
	}
}

func (s *SwitchValueIfEmptySubscriber[T]) Cancel(ctx context.Context) {
	s.cancelled = true
}

func (s *SwitchValueIfEmptySubscriber[T]) OnError(ctx context.Context, err error) {
	s.actual.OnError(ctx, err)
}

func (s *SwitchValueIfEmptySubscriber[T]) OnNext(ctx context.Context, v T) {
	if !s.nextOnce {
		s.nextOnce = true
	}
	s.actual.OnNext(ctx, v)
}

func (s *SwitchValueIfEmptySubscriber[T]) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	s.ctx = ctx
	if old := s.su; old != nil {
		// TODO: check should cancel
		old.Cancel(ctx)
	}
	s.su = su
	if n := s.requested; n > 0 {
		s.Request(ctx, n)
	}
}

func (s *SwitchValueIfEmptySubscriber[T]) OnComplete(ctx context.Context) {
	if s.nextOnce {
		s.actual.OnComplete(ctx)
		return
	}
	s.nextOnce = true

	s.actual.OnNext(ctx, s.alternative)
	s.actual.OnComplete(ctx)
}
