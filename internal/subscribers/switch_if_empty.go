package subscribers

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
)

type SwitchIfEmptySubscriber struct {
	ctx       context.Context
	actual    rs.Subscriber
	other     rs.Publisher
	nextOnce  bool
	cancelled bool
	su        rs.Subscription
	requested int
}

func (s *SwitchIfEmptySubscriber) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
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

func (s *SwitchIfEmptySubscriber) OnNext(v interface{}) {
	if !s.nextOnce {
		s.nextOnce = true
	}
	s.actual.OnNext(v)
}

func (s *SwitchIfEmptySubscriber) OnSubscribe(su rs.Subscription) {
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

func NewSwitchIfEmptySubscriber(alternative rs.Publisher, actual rs.Subscriber) *SwitchIfEmptySubscriber {
	return &SwitchIfEmptySubscriber{
		actual: actual,
		other:  alternative,
	}
}
