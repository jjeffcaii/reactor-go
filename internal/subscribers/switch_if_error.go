package subscribers

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

var (
	_ reactor.Subscriber   = (*SwitchIfErrorSubscriber)(nil)
	_ reactor.Subscription = (*SwitchIfErrorSubscriber)(nil)
)

type FnSwitchIfError = func(error) reactor.RawPublisher

type SwitchIfErrorSubscriber struct {
	mu          sync.Mutex
	ctx         context.Context
	alternative FnSwitchIfError
	actual      reactor.Subscriber
	requested   uint32
	su          reactor.Subscription
	errorCalls  int32
	cancelled   int32
}

func (s *SwitchIfErrorSubscriber) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	atomic.AddUint32(&s.requested, uint32(n))

	s.mu.Lock()
	su := s.su
	s.mu.Unlock()
	if su != nil {
		su.Request(n)
	}
}

func (s *SwitchIfErrorSubscriber) Cancel() {
	// TODO: cancel handling?
	atomic.AddInt32(&s.cancelled, 1)
}

func (s *SwitchIfErrorSubscriber) OnComplete() {
	s.actual.OnComplete()
}

func (s *SwitchIfErrorSubscriber) OnError(err error) {
	if atomic.AddInt32(&s.errorCalls, 1) == 1 {
		// TODO: trigger drop error hook here?
		s.alternative(err).SubscribeWith(s.ctx, s)
	} else {
		s.actual.OnError(err)
	}
}

func (s *SwitchIfErrorSubscriber) OnNext(any reactor.Any) {
	s.actual.OnNext(any)
}

func (s *SwitchIfErrorSubscriber) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	s.mu.Lock()
	s.ctx = ctx
	prevSu := s.su
	s.su = subscription
	s.mu.Unlock()

	if prevSu != nil {
		// TODO: check should cancel
		prevSu.Cancel()
	}

	if n := atomic.LoadUint32(&s.requested); n > 0 {
		s.Request(int(n))
	}
}

func NewSwitchIfErrorSubscriber(alternative FnSwitchIfError, actual reactor.Subscriber) *SwitchIfErrorSubscriber {
	return &SwitchIfErrorSubscriber{
		alternative: alternative,
		actual:      actual,
	}
}
