package subscribers

import (
	"sync"

	rs "github.com/jjeffcaii/reactor-go"
)

type DoFinallySubscriber struct {
	actual    rs.Subscriber
	onFinally rs.FnOnFinally
	once      sync.Once
	s         rs.Subscription
}

func (p *DoFinallySubscriber) Request(n int) {
	p.s.Request(n)
}

func (p *DoFinallySubscriber) Cancel() {
	p.s.Cancel()
	p.runFinally(rs.SignalTypeCancel)
}

func (p *DoFinallySubscriber) OnError(err error) {
	p.actual.OnError(err)
	p.runFinally(rs.SignalTypeError)
}

func (p *DoFinallySubscriber) OnNext(v interface{}) {
	p.actual.OnNext(v)
}

func (p *DoFinallySubscriber) OnSubscribe(s rs.Subscription) {
	p.s = s
	p.actual.OnSubscribe(p)
}

func (p *DoFinallySubscriber) OnComplete() {
	p.actual.OnComplete()
	p.runFinally(rs.SignalTypeComplete)
}

func (p *DoFinallySubscriber) runFinally(sig rs.SignalType) {
	p.once.Do(func() {
		p.onFinally(sig)
	})
}

func NewDoFinallySubscriber(actual rs.Subscriber, onFinally rs.FnOnFinally) *DoFinallySubscriber {
	return &DoFinallySubscriber{
		onFinally: onFinally,
		actual:    actual,
	}
}
