package subscribers

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

var (
	_ reactor.Subscriber[any] = (*SwitchIfErrorSubscriber[any])(nil)
	_ reactor.Subscription    = (*SwitchIfErrorSubscriber[any])(nil)
)

type FnSwitchIfError[T any] func(context.Context, error) reactor.RawPublisher[T]

type SwitchIfErrorSubscriber[T any] struct {
	mu          sync.Mutex
	ctx         context.Context
	alternative FnSwitchIfError[T]
	actual      reactor.Subscriber[T]
	requested   uint32
	su          reactor.Subscription
	errorCalls  int32
	cancelled   int32
}

func NewSwitchIfErrorSubscriber[T any](alternative FnSwitchIfError[T], actual reactor.Subscriber[T]) *SwitchIfErrorSubscriber[T] {
	return &SwitchIfErrorSubscriber[T]{
		alternative: alternative,
		actual:      actual,
	}
}

func (s *SwitchIfErrorSubscriber[T]) Request(ctx context.Context, n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	atomic.AddUint32(&s.requested, uint32(n))

	s.mu.Lock()
	su := s.su
	s.mu.Unlock()
	if su != nil {
		su.Request(ctx, n)
	}
}

func (s *SwitchIfErrorSubscriber[T]) Cancel(ctx context.Context) {
	// TODO: cancel handling?
	atomic.AddInt32(&s.cancelled, 1)
}

func (s *SwitchIfErrorSubscriber[T]) OnComplete(ctx context.Context) {
	s.actual.OnComplete(ctx)
}

func (s *SwitchIfErrorSubscriber[T]) OnError(ctx context.Context, err error) {
	if atomic.AddInt32(&s.errorCalls, 1) == 1 {
		// TODO: trigger drop error hook here?
		s.alternative(ctx, err).SubscribeWith(s.ctx, s)
	} else {
		s.actual.OnError(ctx, err)
	}
}

func (s *SwitchIfErrorSubscriber[T]) OnNext(ctx context.Context, next T) {
	s.actual.OnNext(ctx, next)
}

func (s *SwitchIfErrorSubscriber[T]) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	s.mu.Lock()
	s.ctx = ctx
	prevSu := s.su
	s.su = subscription
	s.mu.Unlock()

	if prevSu != nil {
		// TODO: check should cancel
		prevSu.Cancel(ctx)
	}

	if n := atomic.LoadUint32(&s.requested); n > 0 {
		s.Request(ctx, int(n))
	}
}
