package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
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
	*baseMono
	source Mono
	mapper rs.Transformer
}

func (m *monoMap) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (m *monoMap) SubscribeWith(ctx context.Context, sub rs.Subscriber) {
	m.source.SubscribeWith(ctx, newMapSubscriber(sub, m.mapper))
}

func newMonoMap(source Mono, tf rs.Transformer) Mono {
	m := &monoMap{
		source: source,
		mapper: tf,
	}
	m.baseMono = &baseMono{
		child: m,
	}
	return m
}
