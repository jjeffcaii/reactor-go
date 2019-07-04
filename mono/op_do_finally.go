package mono

import (
	"context"
	"sync"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type monoDoFinally struct {
	source    Mono
	onFinally rs.FnOnFinally
}

func (m monoDoFinally) DoFinally(fn rs.FnOnFinally) Mono {
	return newMonoDoFinally(m, fn)
}

func (m monoDoFinally) Filter(f rs.Predicate) Mono {
	return newMonoFilter(m, f)
}

func (m monoDoFinally) Map(t rs.Transformer) Mono {
	return newMonoMap(m, t)
}

func (m monoDoFinally) FlatMap(mapper flatMapper) Mono {
	return newMonoFlatMap(m, mapper)
}

func (m monoDoFinally) SubscribeOn(sc scheduler.Scheduler) Mono {
	return newMonoScheduleOn(m, sc)
}

func (m monoDoFinally) Block(ctx context.Context) (interface{}, error) {
	return toBlock(ctx, m)
}

func (m monoDoFinally) DoOnNext(fn rs.FnOnNext) Mono {
	return newMonoPeek(m, peekNext(fn))
}

func (m monoDoFinally) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeRaw(ctx, rs.NewSubscriber(options...))
}

func (m monoDoFinally) SubscribeRaw(ctx context.Context, s rs.Subscriber) {
	m.source.SubscribeRaw(ctx, newDoFinallySubscriber(s, m.onFinally))
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
	return monoDoFinally{
		source:    source,
		onFinally: onFinally,
	}
}
