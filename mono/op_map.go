package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type mapSubscriber struct {
	actual reactor.Subscriber
	t      reactor.Transformer
}

func (m mapSubscriber) OnComplete() {
	m.actual.OnComplete()
}

func (m mapSubscriber) OnError(err error) {
	m.actual.OnError(err)
}

func (m mapSubscriber) OnNext(v Any) {
	if o, err := m.t(v); err != nil {
		m.actual.OnError(err)
	} else {
		m.actual.OnNext(o)
	}
}

func (m mapSubscriber) OnSubscribe(s reactor.Subscription) {
	m.actual.OnSubscribe(s)
}

func newMapSubscriber(s reactor.Subscriber, t reactor.Transformer) mapSubscriber {
	return mapSubscriber{
		actual: s,
		t:      t,
	}
}

type monoMap struct {
	source reactor.RawPublisher
	mapper reactor.Transformer
}

func (m *monoMap) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, newMapSubscriber(actual, m.mapper))
	m.source.SubscribeWith(ctx, actual)
}

func newMonoMap(source reactor.RawPublisher, tf reactor.Transformer) *monoMap {
	return &monoMap{
		source: source,
		mapper: tf,
	}
}
