package flux

import (
	"context"

	"github.com/pkg/errors"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type fluxFilter[T any] struct {
	source    reactor.RawPublisher[T]
	predicate reactor.Predicate[T]
}

func newFluxFilter[T any](source reactor.RawPublisher[T], predicate reactor.Predicate[T]) *fluxFilter[T] {
	return &fluxFilter[T]{
		source:    source,
		predicate: predicate,
	}
}

func (f *fluxFilter[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	// TODO: fuse
	sub := newFilterSubscriber(s, f.predicate)
	f.source.SubscribeWith(ctx, sub)
}

type filterSubscriber[T any] struct {
	ctx    context.Context
	actual reactor.Subscriber[T]
	f      reactor.Predicate[T]
	su     reactor.Subscription
}

func newFilterSubscriber[T any](s reactor.Subscriber[T], p reactor.Predicate[T]) *filterSubscriber[T] {
	return &filterSubscriber[T]{
		actual: s,
		f:      p,
	}
}

func (p *filterSubscriber[T]) OnComplete(ctx context.Context) {
	p.actual.OnComplete(ctx)
}

func (p *filterSubscriber[T]) OnError(ctx context.Context, err error) {
	p.actual.OnError(ctx, err)
}

func (p *filterSubscriber[T]) OnNext(ctx context.Context, v T) {
	if p.f == nil {
		panic("the Filter predicate is nil")
		return
	}

	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			p.OnError(ctx, errors.WithStack(e))
		} else {
			p.OnError(ctx, errors.Errorf("%v", rec))
		}
	}()

	if p.f(ctx, v) {
		p.actual.OnNext(ctx, v)
		return
	}
	p.su.Request(ctx, 1)
	internal.TryDiscard(p.ctx, v)
}

func (p *filterSubscriber[T]) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	p.ctx = ctx
	p.su = su
	p.actual.OnSubscribe(ctx, su)
}
