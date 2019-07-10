package mono

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type filterSubscriber struct {
	*internal.ContextSupport
	actual rs.Subscriber
	f      rs.Predicate
}

func (f filterSubscriber) OnComplete() {
	f.actual.OnComplete()
}

func (f filterSubscriber) OnError(err error) {
	f.actual.OnError(err)
}

func (f filterSubscriber) OnNext(v interface{}) {
	if f.f(v) {
		f.actual.OnNext(v)
		return
	}
	internal.TryDiscard(f.ContextSupport, v)
}

func (f filterSubscriber) OnSubscribe(s rs.Subscription) {
	f.actual.OnSubscribe(s)
}

type monoFilter struct {
	s Mono
	f rs.Predicate
}

func (m *monoFilter) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	m.s.SubscribeWith(ctx, newFilterSubscriber(ctx, s, m.f))
}

func newFilterSubscriber(ctx context.Context, actual rs.Subscriber, predicate rs.Predicate) filterSubscriber {
	return filterSubscriber{
		ContextSupport: internal.NewContextSupport(ctx),
		actual:         actual,
		f:              predicate,
	}
}

func newMonoFilter(s Mono, f rs.Predicate) *monoFilter {
	return &monoFilter{
		s: s,
		f: f,
	}
}
