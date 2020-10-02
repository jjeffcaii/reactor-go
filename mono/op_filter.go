package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type filterSubscriber struct {
	ctx       context.Context
	actual    reactor.Subscriber
	predicate reactor.Predicate
}

func (f *filterSubscriber) OnComplete() {
	f.actual.OnComplete()
}

func (f *filterSubscriber) OnError(err error) {
	f.actual.OnError(err)
}

func (f *filterSubscriber) OnNext(v Any) {
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			f.OnError(err)
		}
	}()
	if f.predicate(v) {
		f.actual.OnNext(v)
		return
	}
	internal.TryDiscard(f.ctx, v)
}

func (f *filterSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	f.ctx = ctx
	f.actual.OnSubscribe(ctx, s)
}

type monoFilter struct {
	s reactor.RawPublisher
	f reactor.Predicate
}

func (m *monoFilter) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	m.s.SubscribeWith(ctx, newFilterSubscriber(s, m.f))
}

func newFilterSubscriber(actual reactor.Subscriber, predicate reactor.Predicate) *filterSubscriber {
	return &filterSubscriber{
		actual:    actual,
		predicate: predicate,
	}
}

func newMonoFilter(s reactor.RawPublisher, f reactor.Predicate) *monoFilter {
	return &monoFilter{
		s: s,
		f: f,
	}
}
