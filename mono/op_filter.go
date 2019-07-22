package mono

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type filterSubscriber struct {
	ctx    context.Context
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
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			f.OnError(err)
		}
	}()
	if f.f(v) {
		f.actual.OnNext(v)
		return
	}
	internal.TryDiscard(f.ctx, v)
}

func (f filterSubscriber) OnSubscribe(s rs.Subscription) {
	f.actual.OnSubscribe(s)
}

type monoFilter struct {
	s rs.RawPublisher
	f rs.Predicate
}

func (m *monoFilter) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, newFilterSubscriber(ctx, actual, m.f))
	m.s.SubscribeWith(ctx, actual)
}

func newFilterSubscriber(ctx context.Context, actual rs.Subscriber, predicate rs.Predicate) filterSubscriber {
	return filterSubscriber{
		ctx:    ctx,
		actual: actual,
		f:      predicate,
	}
}

func newMonoFilter(s rs.RawPublisher, f rs.Predicate) *monoFilter {
	return &monoFilter{
		s: s,
		f: f,
	}
}
