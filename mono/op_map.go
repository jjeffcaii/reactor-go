package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type mapSubscriber struct {
	actual rs.Subscriber
	t      rs.Transformer
}

func (m mapSubscriber) OnComplete() {
	m.actual.OnComplete()
}

func (m mapSubscriber) OnError(err error) {
	m.actual.OnError(err)
}

func (m mapSubscriber) OnNext(v interface{}) {
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			m.OnError(err)
		}
	}()
	m.actual.OnNext(m.t(v))
}

func (m mapSubscriber) OnSubscribe(s rs.Subscription) {
	m.actual.OnSubscribe(s)
}

func newMapSubscriber(s rs.Subscriber, t rs.Transformer) mapSubscriber {
	return mapSubscriber{
		actual: s,
		t:      t,
	}
}

type monoMap struct {
	source rs.RawPublisher
	mapper rs.Transformer
}

func (m *monoMap) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, newMapSubscriber(actual, m.mapper))
	m.source.SubscribeWith(ctx, actual)
}

func newMonoMap(source rs.RawPublisher, tf rs.Transformer) *monoMap {
	return &monoMap{
		source: source,
		mapper: tf,
	}
}
