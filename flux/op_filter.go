package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type filterSubscriber struct {
	ctx    context.Context
	actual reactor.Subscriber
	f      reactor.Predicate
	su     reactor.Subscription
}

func (p *filterSubscriber) OnComplete() {
	p.actual.OnComplete()
}

func (p *filterSubscriber) OnError(err error) {
	p.actual.OnError(err)
}

func (p *filterSubscriber) OnNext(v Any) {
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

func (p *filterSubscriber) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	p.ctx = ctx
	p.su = su
	p.actual.OnSubscribe(ctx, su)
}

func newFilterSubscriber(s reactor.Subscriber, p reactor.Predicate) *filterSubscriber {
	return &filterSubscriber{
		actual: s,
		f:      p,
	}
}

type fluxFilter struct {
	source    reactor.RawPublisher
	predicate reactor.Predicate
}

func (f *fluxFilter) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	var actual reactor.Subscriber
	if cs, ok := s.(*internal.CoreSubscriber); ok {
		cs.Subscriber = newFilterSubscriber(cs.Subscriber, f.predicate)
		actual = cs
	} else {
		actual = internal.NewCoreSubscriber(newFilterSubscriber(s, f.predicate))
	}
	f.source.SubscribeWith(ctx, actual)
}

func newFluxFilter(source reactor.RawPublisher, predicate reactor.Predicate) *fluxFilter {
	return &fluxFilter{
		source:    source,
		predicate: predicate,
	}
}
