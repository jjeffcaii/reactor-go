package subscribers

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

type SwitchIfEmptySubscriber struct {
	ctx       context.Context
	actual    reactor.Subscriber
	other     reactor.RawPublisher
	nextOnce  bool
	cancelled bool
	su        reactor.Subscription
	requested int
}

func (s *SwitchIfEmptySubscriber) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	s.requested = n
	if s.su != nil {
		s.su.Request(n)
	}
}

func (s *SwitchIfEmptySubscriber) Cancel() {
	s.cancelled = true
}

func (s *SwitchIfEmptySubscriber) OnError(err error) {
	s.actual.OnError(err)
}

func (s *SwitchIfEmptySubscriber) OnNext(v reactor.Any) {
	if !s.nextOnce {
		s.nextOnce = true
	}
	s.actual.OnNext(v)
}

func (s *SwitchIfEmptySubscriber) OnSubscribe(su reactor.Subscription) {
	if old := s.su; old != nil {
		// TODO: check should cancel
		old.Cancel()
	}
	s.su = su
	if n := s.requested; n > 0 {
		s.Request(n)
	}
}

func (s *SwitchIfEmptySubscriber) OnComplete() {
	if !s.nextOnce {
		s.nextOnce = true
		s.other.SubscribeWith(s.ctx, s)
	} else {
		s.actual.OnComplete()
	}
}

func NewSwitchIfEmptySubscriber(ctx context.Context, alternative reactor.RawPublisher, actual reactor.Subscriber) *SwitchIfEmptySubscriber {
	return &SwitchIfEmptySubscriber{
		ctx:    ctx,
		actual: actual,
		other:  alternative,
	}
}
