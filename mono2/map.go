package mono

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/jjeffcaii/reactor-go"
)

type monoMap[A, B any] struct {
	source reactor.RawPublisher[A]
	mapper reactor.Transformer[A, B]
}

func newMonoMap[A, B any](source reactor.RawPublisher[A], tf reactor.Transformer[A, B]) monoMap[A, B] {
	return monoMap[A, B]{
		source: source,
		mapper: tf,
	}
}

func (m monoMap[A, B]) SubscribeWith(ctx context.Context, s reactor.Subscriber[B]) {
	suu := &mapSubscriber[A, B]{
		actual: s,
		t:      m.mapper,
	}
	m.source.SubscribeWith(ctx, suu)
}

type mapSubscriber[A, B any] struct {
	actual reactor.Subscriber[B]
	t      reactor.Transformer[A, B]
	stat   int32
}

func (m *mapSubscriber[A, B]) OnComplete(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&m.stat, 0, statComplete) {
		m.actual.OnComplete(ctx)
	}
}

func (m *mapSubscriber[A, B]) OnError(ctx context.Context, err error) {
	if atomic.CompareAndSwapInt32(&m.stat, 0, statError) {
		m.actual.OnError(ctx, err)
	}
}

func (m *mapSubscriber[A, B]) OnNext(ctx context.Context, v A) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			m.OnError(ctx, errors.WithStack(e))
		} else {
			m.OnError(ctx, errors.Errorf("%v", rec))
		}
	}()

	if newValue, err := m.t(ctx, v); err != nil {
		m.actual.OnError(ctx, err)
	} else {
		m.actual.OnNext(ctx, newValue)
	}
}

func (m *mapSubscriber[A, B]) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	m.actual.OnSubscribe(ctx, s)
}
