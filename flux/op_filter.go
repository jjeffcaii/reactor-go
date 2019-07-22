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
	su     rs.Subscription
}

func (p *filterSubscriber) OnComplete() {
	p.actual.OnComplete()
}

func (p *filterSubscriber) OnError(err error) {
	p.actual.OnError(err)
}

func (p *filterSubscriber) OnNext(v interface{}) {
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			p.OnError(err)
		}
	}()
	if p.f(v) {
		p.actual.OnNext(v)
		return
	}
	p.su.Request(1)
	internal.TryDiscard(p.ctx, v)
}

func (p *filterSubscriber) OnSubscribe(su rs.Subscription) {
	p.su = su
	p.actual.OnSubscribe(su)
}

func newFilterSubscriber(ctx context.Context, s rs.Subscriber, p rs.Predicate) *filterSubscriber {
	return &filterSubscriber{
		ctx:    ctx,
		actual: s,
		f:      p,
	}
}

type fluxFilter struct {
	source    rs.RawPublisher
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

func newFluxFilter(source rs.RawPublisher, predicate rs.Predicate) *fluxFilter {
	return &fluxFilter{
		source:    source,
		predicate: predicate,
	}
}
