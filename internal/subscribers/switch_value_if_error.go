package subscribers

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

var (
	_ reactor.Subscription = (*SwitchValueIfErrorSubscriber)(nil)
	_ reactor.Subscriber   = (*SwitchValueIfErrorSubscriber)(nil)
)

type SwitchValueIfErrorSubscriber struct {
	actual     reactor.Subscriber
	v          reactor.Any
	su         reactor.Subscription
	done       int32
	errorCalls int32
}

func NewSwitchValueIfErrorSubscriber(actual reactor.Subscriber, value reactor.Any) *SwitchValueIfErrorSubscriber {
	return &SwitchValueIfErrorSubscriber{
		actual: actual,
		v:      value,
	}
}

func (s *SwitchValueIfErrorSubscriber) Request(n int) {
	if atomic.LoadInt32(&s.done) == 0 && s.su != nil {
		s.su.Request(n)
	}
}

func (s *SwitchValueIfErrorSubscriber) Cancel() {
	s.su.Cancel()
}

func (s *SwitchValueIfErrorSubscriber) OnError(err error) {
	if atomic.AddInt32(&s.errorCalls, 1) == 1 {
		hooks.Global().OnErrorDrop(err)
		s.actual.OnNext(s.v)
		s.actual.OnComplete()
	} else {
		s.actual.OnError(err)
	}
}

func (s *SwitchValueIfErrorSubscriber) OnNext(v reactor.Any) {
	s.actual.OnNext(v)
}

func (s *SwitchValueIfErrorSubscriber) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	s.su = su
	s.actual.OnSubscribe(ctx, s)
}

func (s *SwitchValueIfErrorSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&s.done, 0, 1) {
		s.actual.OnComplete()
	}
}
