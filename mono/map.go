package mono

import (
	"context"
	"math"
	"sync"
	"sync/atomic"

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
		exist.stat = 0
		return exist
	}
	return &mapSubscriber{}
}

func (p *mapSubscriberPool) put(s *mapSubscriber) {
	if s == nil {
		return
	}
	atomic.StoreInt32(&s.stat, math.MinInt32)
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
	stat   int32
}

func (m *mapSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&m.stat, 0, statComplete) {
		defer globalMapSubscriberPool.put(m)
		m.actual.OnComplete()
	}
}

func (m *mapSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&m.stat, 0, statError) {
		defer globalMapSubscriberPool.put(m)
		m.actual.OnError(err)
	}
}

func (m *mapSubscriber) OnNext(v Any) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			m.OnError(errors.WithStack(e))
		} else {
			m.OnError(errors.Errorf("%v", rec))
		}
	}()

	if newValue, err := m.t(v); err != nil {
		m.actual.OnError(err)
	} else {
		m.actual.OnNext(newValue)
	}
}

func (m *mapSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	m.actual.OnSubscribe(ctx, s)
}
