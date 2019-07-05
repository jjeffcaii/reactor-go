package flux

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
)

type filterSubscriber struct {
	source    rs.Subscriber
	predicate rs.Predicate
}

func (s filterSubscriber) OnComplete() {
	s.source.OnComplete()
}

func (s filterSubscriber) OnError(err error) {
	s.source.OnError(err)
}

func (s filterSubscriber) OnNext(su rs.Subscription, in interface{}) {
	if s.predicate(in) {
		s.source.OnNext(su, in)
	}
}

func (s filterSubscriber) OnSubscribe(ss rs.Subscription) {
	s.source.OnSubscribe(ss)
}

func newFilterSubscriber(s rs.Subscriber, p rs.Predicate) rs.Subscriber {
	return filterSubscriber{
		source:    s,
		predicate: p,
	}
}

type fluxFilter struct {
	source    Flux
	predicate rs.Predicate
}

func (f *fluxFilter) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	f.source.SubscribeWith(ctx, newFilterSubscriber(s, f.predicate))
}

func newFluxFilter(source Flux, predicate rs.Predicate) *fluxFilter {
	return &fluxFilter{
		source:    source,
		predicate: predicate,
	}
}
