package mono

import (
	"context"
	"sync"

	rs "github.com/jjeffcaii/reactor-go"
)

type monoDoFinally struct {
	*baseMono
	source    Mono
	onFinally rs.FnOnFinally
}

func (m *monoDoFinally) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (m *monoDoFinally) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	m.source.SubscribeWith(ctx, newDoFinallySubscriber(s, m.onFinally))
}

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

func (p *doFinallySubscriber) OnNext(s rs.Subscription, v interface{}) {
	p.actual.OnNext(s, v)
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

func newMonoDoFinally(source Mono, onFinally rs.FnOnFinally) Mono {
	m := &monoDoFinally{
		source:    source,
		onFinally: onFinally,
	}
	m.baseMono = &baseMono{
		child: m,
	}
	return m
}
