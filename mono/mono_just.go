package mono

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

var _justSubscriptionPool = sync.Pool{
	New: func() interface{} {
		return new(justSubscription)
	},
}

type monoJust struct {
	value Any
}

type justSubscription struct {
	actual reactor.Subscriber
	parent *monoJust
	n      int32
}

func newMonoJust(v Any) *monoJust {
	return &monoJust{
		value: v,
	}
}

func borrowJustSubscription() *justSubscription {
	return _justSubscriptionPool.Get().(*justSubscription)
}

func returnJustSubscription(s *justSubscription) {
	if s == nil {
		return
	}
	s.actual = nil
	s.parent = nil
	atomic.StoreInt32(&s.n, 0)
	_justSubscriptionPool.Put(s)
}

func (j *justSubscription) Request(n int) {
	if j == nil || j.actual == nil || j.parent == nil {
		return
	}
	if n < 1 {
		returnJustSubscription(j)
		panic(reactor.ErrNegativeRequest)
	}
	if !atomic.CompareAndSwapInt32(&j.n, 0, statComplete) {
		return
	}

	defer returnJustSubscription(j)

	if j.parent.value == nil {
		j.actual.OnComplete()
		return
	}
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			j.actual.OnError(err)
		} else {
			j.actual.OnComplete()
		}
	}()
	j.actual.OnNext(j.parent.value)
}

func (j *justSubscription) Cancel() {
	if j == nil || j.parent == nil || j.actual == nil {
		return
	}
	atomic.CompareAndSwapInt32(&j.n, 0, statCancel)
}

func (m *monoJust) SubscribeWith(ctx context.Context, sub reactor.Subscriber) {
	select {
	case <-ctx.Done():
		sub.OnError(reactor.ErrSubscribeCancelled)
	default:
		su := borrowJustSubscription()
		su.parent = m
		su.actual = sub
		sub.OnSubscribe(ctx, su)
	}
}
