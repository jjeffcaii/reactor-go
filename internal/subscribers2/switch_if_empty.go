package subscribers

import (
	"context"
	"errors"

	"github.com/jjeffcaii/reactor-go"
)

var errNilAlternativeMono = errors.New("the alternative SwitchIfEmpty Mono is nil")

var (
	_ reactor.Subscriber[any] = (*SwitchIfEmptySubscriber[any])(nil)
	_ reactor.Subscription    = (*SwitchIfEmptySubscriber[any])(nil)
)

type SwitchIfEmptySubscriber[T any] struct {
	ctx       context.Context
	actual    reactor.Subscriber[T]
	other     reactor.RawPublisher[T]
	nextOnce  bool
	cancelled bool
	su        reactor.Subscription
	requested int
}

func NewSwitchIfEmptySubscriber[T any](alternative reactor.RawPublisher[T], actual reactor.Subscriber[T]) *SwitchIfEmptySubscriber[T] {
	return &SwitchIfEmptySubscriber[T]{
		actual: actual,
		other:  alternative,
	}
}

func (s *SwitchIfEmptySubscriber[T]) Request(ctx context.Context, n int) {
	if n < 1 {
		return
	}
	s.requested = n
	if s.su != nil {
		s.su.Request(ctx, n)
	}
}

func (s *SwitchIfEmptySubscriber[T]) Cancel(ctx context.Context) {
	s.cancelled = true
}

func (s *SwitchIfEmptySubscriber[T]) OnError(ctx context.Context, err error) {
	s.actual.OnError(ctx, err)
}

func (s *SwitchIfEmptySubscriber[T]) OnNext(ctx context.Context, v T) {
	if !s.nextOnce {
		s.nextOnce = true
	}
	s.actual.OnNext(ctx, v)
}

func (s *SwitchIfEmptySubscriber[T]) OnSubscribe(ctx context.Context, su reactor.Subscription) {
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

func (s *SwitchIfEmptySubscriber[T]) OnComplete(ctx context.Context) {
	if s.nextOnce {
		s.actual.OnComplete(ctx)
		return
	}
	s.nextOnce = true
	if s.other == nil {
		s.actual.OnError(ctx, errNilAlternativeMono)
	} else {
		s.other.SubscribeWith(s.ctx, s)
	}
}
