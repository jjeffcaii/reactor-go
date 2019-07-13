package flux

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

func (s filterSubscriber) OnComplete() {
	s.actual.OnComplete()
}

func (s filterSubscriber) OnError(err error) {
	s.actual.OnError(err)
}

func (s filterSubscriber) OnNext(v interface{}) {
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			s.OnError(err)
		}
	}()
	if s.f(v) {
		s.actual.OnNext(v)
		return
	}
	internal.TryDiscard(s.ctx, v)
}

func (s filterSubscriber) OnSubscribe(ss rs.Subscription) {
	s.actual.OnSubscribe(ss)
}

func newFilterSubscriber(ctx context.Context, s rs.Subscriber, p rs.Predicate) rs.Subscriber {
	return filterSubscriber{
		ctx:    ctx,
		actual: s,
		f:      p,
	}
}

type fluxFilter struct {
	source    Flux
	predicate rs.Predicate
}

func (f *fluxFilter) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	var actual rs.Subscriber
	if cs, ok := s.(*internal.CoreSubscriber); ok {
		cs.Subscriber = newFilterSubscriber(ctx, cs.Subscriber, f.predicate)
		actual = cs
	} else {
		actual = internal.NewCoreSubscriber(ctx, newFilterSubscriber(ctx, s, f.predicate))
	}
	f.source.SubscribeWith(ctx, actual)
}

func newFluxFilter(source Flux, predicate rs.Predicate) *fluxFilter {
	return &fluxFilter{
		source:    source,
		predicate: predicate,
	}
}
