package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type mapSubscriber struct {
	s rs.Subscriber
	t rs.Transformer
}

func (m mapSubscriber) OnComplete() {
	m.s.OnComplete()
}

func (m mapSubscriber) OnError(err error) {
	m.s.OnError(err)
}

func (m mapSubscriber) OnNext(s rs.Subscription, v interface{}) {
	m.s.OnNext(s, m.t(v))
}

func (m mapSubscriber) OnSubscribe(s rs.Subscription) {
	m.s.OnSubscribe(s)
}

func newMapSubscriber(s rs.Subscriber, t rs.Transformer) mapSubscriber {
	return mapSubscriber{
		s: s,
		t: t,
	}
}

type monoMap struct {
	source Mono
	mapper rs.Transformer
}

func (m monoMap) DoFinally(fn rs.FnOnFinally) Mono {
	return newMonoDoFinally(m, fn)
}

func (m monoMap) DoOnNext(fn rs.FnOnNext) Mono {
	return newMonoPeek(m, peekNext(fn))
}

func (m monoMap) Block(ctx context.Context) (interface{}, error) {
	return toBlock(ctx, m)
}

func (m monoMap) FlatMap(f flatMapper) Mono {
	return newMonoFlatMap(m, f)
}

func (m monoMap) SubscribeOn(sc scheduler.Scheduler) Mono {
	return newMonoScheduleOn(m, sc)
}

func (m monoMap) Filter(f rs.Predicate) Mono {
	return newMonoFilter(m, f)
}

func (m monoMap) Map(t rs.Transformer) Mono {
	return newMonoMap(m, t)
}

func (m monoMap) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeRaw(ctx, rs.NewSubscriber(options...))
}

func (m monoMap) SubscribeRaw(ctx context.Context, sub rs.Subscriber) {
	m.source.SubscribeRaw(ctx, newMapSubscriber(sub, m.mapper))
}

func newMonoMap(source Mono, tf rs.Transformer) Mono {
	return monoMap{
		source: source,
		mapper: tf,
	}
}
