package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/pkg/errors"
)

type filterSubscriber struct {
	ctx       context.Context
	actual    reactor.Subscriber
	predicate reactor.Predicate
}

type monoFilter struct {
	s reactor.RawPublisher
	f reactor.Predicate
}

func newFilterSubscriber(actual reactor.Subscriber, predicate reactor.Predicate) *filterSubscriber {
	return &filterSubscriber{
		actual:    actual,
		predicate: predicate,
	}
}

func newMonoFilter(s reactor.RawPublisher, f reactor.Predicate) monoFilter {
	return monoFilter{
		s: s,
		f: f,
	}
}

func (f *filterSubscriber) OnComplete() {
	f.actual.OnComplete()
}

func (f *filterSubscriber) OnError(err error) {
	f.actual.OnError(err)
}

func (f *filterSubscriber) OnNext(v Any) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			f.OnError(errors.WithStack(e))
		} else {
			f.OnError(errors.Errorf("%v", rec))
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

func (m monoFilter) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	m.s.SubscribeWith(ctx, newFilterSubscriber(s, m.f))
}
