package mono

import (
	"context"
	"sync"

	"github.com/jjeffcaii/reactor-go"
	"github.com/pkg/errors"
)

var _mapSubscriberPool = sync.Pool{
	New: func() interface{} {
		return new(mapSubscriber)
	},
}

type mapSubscriber struct {
	actual reactor.Subscriber
	t      reactor.Transformer
}

type monoMap struct {
	source reactor.RawPublisher
	mapper reactor.Transformer
}

func (m monoMap) Parent() reactor.RawPublisher {
	return m.source
}

func newMonoMap(source reactor.RawPublisher, tf reactor.Transformer) monoMap {
	return monoMap{
		source: source,
		mapper: tf,
	}
}

func borrowMapSubscriber(s reactor.Subscriber, t reactor.Transformer) *mapSubscriber {
	sub := _mapSubscriberPool.Get().(*mapSubscriber)
	sub.actual = s
	sub.t = t
	return sub
}

func returnMapSubscriber(s *mapSubscriber) {
	if s == nil {
		return
	}
	s.actual = nil
	s.t = nil
	_mapSubscriberPool.Put(s)
}

func (m *mapSubscriber) OnComplete() {
	if m == nil || m.actual == nil || m.t == nil {
		return
	}
	defer returnMapSubscriber(m)
	m.actual.OnComplete()
}

func (m *mapSubscriber) OnError(err error) {
	if m == nil || m.actual == nil || m.t == nil {
		return
	}
	defer returnMapSubscriber(m)
	m.actual.OnError(err)
}

func (m *mapSubscriber) OnNext(v Any) {
	if m == nil || m.actual == nil || m.t == nil {
		// TODO:
		return
	}
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			m.actual.OnError(errors.WithStack(e))
		} else {
			m.actual.OnError(errors.Errorf("%v", rec))
		}
	}()

	if transformed, err := m.t(v); err != nil {
		m.actual.OnError(err)
	} else {
		m.actual.OnNext(transformed)
	}
}

func (m *mapSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	if m == nil || m.actual == nil || m.t == nil {
		return
	}
	m.actual.OnSubscribe(ctx, s)
}

func (m monoMap) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	m.source.SubscribeWith(ctx, borrowMapSubscriber(s, m.mapper))
}
