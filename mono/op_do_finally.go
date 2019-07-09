package mono

import (
	"context"
	"sync"

	rs "github.com/jjeffcaii/reactor-go"
)

type doFinallySubscriber struct {
	actual    rs.Subscriber
	onFinally rs.FnOnFinally
	once      sync.Once
	s         rs.Subscription
}

func (p *doFinallySubscriber) Request(n int) {
	p.s.Request(n)
}

func (p *doFinallySubscriber) Cancel() {
	p.s.Cancel()
	p.runFinally(rs.SignalCancel)
}

func (p *doFinallySubscriber) OnError(err error) {
	p.actual.OnError(err)
	p.runFinally(rs.SignalError)
}

func (p *doFinallySubscriber) OnNext(v interface{}) {
	p.actual.OnNext(v)
}

func (p *doFinallySubscriber) OnSubscribe(s rs.Subscription) {
	p.s = s
	p.actual.OnSubscribe(p)
}

func (p *doFinallySubscriber) OnComplete() {
	p.actual.OnComplete()
	p.runFinally(rs.SignalComplete)
}

func (p *doFinallySubscriber) runFinally(sig rs.Signal) {
	p.once.Do(func() {
		p.onFinally(sig)
	})
}

func newDoFinallySubscriber(actual rs.Subscriber, onFinally rs.FnOnFinally) *doFinallySubscriber {
	return &doFinallySubscriber{
		onFinally: onFinally,
		actual:    actual,
	}
}

type monoDoFinally struct {
	source    Mono
	onFinally rs.FnOnFinally
}

func (m *monoDoFinally) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	m.source.SubscribeWith(ctx, newDoFinallySubscriber(s, m.onFinally))
}

func newMonoDoFinally(source Mono, onFinally rs.FnOnFinally) *monoDoFinally {
	return &monoDoFinally{
		source:    source,
		onFinally: onFinally,
	}
}
