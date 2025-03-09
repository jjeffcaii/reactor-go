package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

type mapSubscriber[A, B any] struct {
	actual reactor.Subscriber[B]
	mapper reactor.Transformer[A, B]
}

func newMapSubscriber[A, B any](s reactor.Subscriber[B], mapper reactor.Transformer[A, B]) mapSubscriber[A, B] {
	return mapSubscriber[A, B]{
		actual: s,
		mapper: mapper,
	}
}

func (s mapSubscriber[A, B]) OnComplete(ctx context.Context) {
	s.actual.OnComplete(ctx)
}

func (s mapSubscriber[A, B]) OnError(ctx context.Context, err error) {
	s.actual.OnError(ctx, err)
}

func (s mapSubscriber[A, B]) OnNext(ctx context.Context, next A) {
	if o, err := s.mapper(ctx, next); err != nil {
		s.actual.OnError(ctx, err)
	} else {
		s.actual.OnNext(ctx, o)
	}
}

func (s mapSubscriber[A, B]) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	s.actual.OnSubscribe(ctx, su)
}

type fluxMap[A, B any] struct {
	source reactor.RawPublisher[A]
	mapper reactor.Transformer[A, B]
}

func newFluxMap[A, B any](source reactor.RawPublisher[A], mapper reactor.Transformer[A, B]) *fluxMap[A, B] {
	return &fluxMap[A, B]{
		source: source,
		mapper: mapper,
	}
}

func (p *fluxMap[A, B]) SubscribeWith(ctx context.Context, sub reactor.Subscriber[B]) {
	p.source.SubscribeWith(ctx, newMapSubscriber(sub, p.mapper))
}
