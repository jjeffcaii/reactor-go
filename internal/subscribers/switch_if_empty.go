package subscribers

import (
	"context"
	"errors"

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

func (s *SwitchIfEmptySubscriber) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	s.ctx = ctx
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
	if s.nextOnce {
		s.actual.OnComplete()
		return
	}
	s.nextOnce = true
	if s.other == nil {
		s.actual.OnError(errors.New("the alternative SwitchIfEmpty Mono is nil"))
	} else {
		s.other.SubscribeWith(s.ctx, s)
	}

}

func NewSwitchIfEmptySubscriber(alternative reactor.RawPublisher, actual reactor.Subscriber) *SwitchIfEmptySubscriber {
	return &SwitchIfEmptySubscriber{
		actual: actual,
		other:  alternative,
	}
}
