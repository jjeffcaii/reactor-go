package mono

import (
	"context"
	"sync"

	"github.com/jjeffcaii/reactor-go"
	"github.com/pkg/errors"
)

var globalMapSubscriberPool mapSubscriberPool

func newMonoMap(source reactor.RawPublisher, tf reactor.Transformer) monoMap {
	return monoMap{
		source: source,
		mapper: tf,
	}
}

type mapSubscriberPool struct {
	inner sync.Pool
}

func (p *mapSubscriberPool) get() *mapSubscriber {
	if exist, _ := p.inner.Get().(*mapSubscriber); exist != nil {
		return exist
	}
	return &mapSubscriber{}
}

func (p *mapSubscriberPool) put(s *mapSubscriber) {
	if s == nil {
		return
	}
	s.actual = nil
	s.t = nil
	p.inner.Put(s)
}

type monoMap struct {
	source reactor.RawPublisher
	mapper reactor.Transformer
}

func (m monoMap) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	su := globalMapSubscriberPool.get()
	su.actual = s
	su.t = m.mapper
	m.source.SubscribeWith(ctx, su)
}

type mapSubscriber struct {
	actual reactor.Subscriber
	t      reactor.Transformer
}

func (m *mapSubscriber) OnComplete() {
	if m == nil || m.actual == nil || m.t == nil {
		return
	}
	defer m.actual.OnComplete()
}

func (m *mapSubscriber) OnError(err error) {
	if m == nil || m.actual == nil || m.t == nil {
		return
	}
	defer globalMapSubscriberPool.put(m)
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
