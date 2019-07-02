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

func (mapSubscriber) Raw() rs.RawSubscriber {
	panic("implement me")
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

func (m monoMap) Map(t rs.Transformer) Mono {
	return newMonoMap(m, t)
}

func (m monoMap) Subscribe(ctx context.Context, sub rs.Subscriber) rs.Disposable {
	m.source.Subscribe(ctx, newMapSubscriber(sub, m.mapper))
	return nil
}

func newMonoMap(source Mono, tf rs.Transformer) Mono {
	return monoMap{
		source: source,
		mapper: tf,
	}
}
