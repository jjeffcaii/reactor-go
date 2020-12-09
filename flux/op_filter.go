package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/pkg/errors"
)

type fluxFilter struct {
	source    reactor.RawPublisher
	predicate reactor.Predicate
}

type filterSubscriber struct {
	ctx    context.Context
	actual reactor.Subscriber
	f      reactor.Predicate
	su     reactor.Subscription
}

func newFluxFilter(source reactor.RawPublisher, predicate reactor.Predicate) *fluxFilter {
	return &fluxFilter{
		source:    source,
		predicate: predicate,
	}
}

func newFilterSubscriber(s reactor.Subscriber, p reactor.Predicate) *filterSubscriber {
	return &filterSubscriber{
		actual: s,
		f:      p,
	}
}

func (p *filterSubscriber) OnComplete() {
	p.actual.OnComplete()
}

func (p *filterSubscriber) OnError(err error) {
	p.actual.OnError(err)
}

func (p *filterSubscriber) OnNext(v Any) {
	if p.f == nil {
		p.OnError(errors.New("the Filter predicate is nil"))
		return
	}

	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			p.OnError(errors.WithStack(e))
		} else {
			p.OnError(errors.Errorf("%v", rec))
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

func (f *fluxFilter) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	// TODO: fuse
	sub := newFilterSubscriber(s, f.predicate)
	f.source.SubscribeWith(ctx, sub)
}
