package mono

import (
	"context"
	"math"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/pkg/errors"
)

var globalJustSubscriptionPool justSubscriptionPool

type justSubscriptionPool struct {
	inner sync.Pool
}

func (j *justSubscriptionPool) get() *justSubscription {
	if exist, _ := j.inner.Get().(*justSubscription); exist != nil {
		exist.stat = 0
		return exist
	}
	return &justSubscription{}
}

func (j *justSubscriptionPool) put(s *justSubscription) {
	if s == nil {
		return
	}
	s.actual = nil
	s.parent = nil
	atomic.StoreInt32(&s.stat, math.MinInt32)
	j.inner.Put(s)
}

type monoJust struct {
	value Any
}

func newMonoJust(v Any) *monoJust {
	return &monoJust{
		value: v,
	}
}

type justSubscription struct {
	actual reactor.Subscriber
	parent *monoJust
	stat   int32
}

func (j *justSubscription) Request(n int) {
	defer globalJustSubscriptionPool.put(j)

	if n < 1 {
		j.actual.OnError(errors.Errorf("positive request amount required but it was %d", n))
		return
	}

	if !atomic.CompareAndSwapInt32(&j.stat, 0, statComplete) {
		return
	}

	defer func() {
		actual := j.actual

		rec := recover()
		if rec == nil {
			actual.OnComplete()
			return
		}
		if e, ok := rec.(error); ok {
			actual.OnError(errors.WithStack(e))
		} else {
			actual.OnError(errors.Errorf("%v", rec))
		}
	}()

	if j.parent.value != nil {
		j.actual.OnNext(j.parent.value)
	}
}

func (j *justSubscription) Cancel() {
	if atomic.CompareAndSwapInt32(&j.stat, 0, statCancel) {
		j.actual.OnError(reactor.ErrSubscribeCancelled)
	}
}

func (m *monoJust) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	select {
	case <-ctx.Done():
		actual.OnError(reactor.NewContextError(ctx.Err()))
	default:
		su := globalJustSubscriptionPool.get()
		su.parent = m
		su.actual = actual
		actual.OnSubscribe(ctx, su)
	}
}
