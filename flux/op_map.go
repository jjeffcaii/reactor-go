package flux

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
)

type mapSubscriber struct {
	source rs.Subscriber
	mapper rs.Transformer
}

func (s mapSubscriber) OnComplete() {
	s.source.OnComplete()
}

func (s mapSubscriber) OnError(err error) {
	s.source.OnError(err)
}

func (s mapSubscriber) OnNext(su rs.Subscription, v interface{}) {
	s.source.OnNext(su, s.mapper(v))
}

func (s mapSubscriber) OnSubscribe(subscription rs.Subscription) {
	s.source.OnSubscribe(subscription)
}

func newMapSubscriber(s rs.Subscriber, mapper rs.Transformer) rs.Subscriber {
	return mapSubscriber{
		source: s,
		mapper: mapper,
	}
}

type fluxMap struct {
	source Flux
	mapper rs.Transformer
}

func (p *fluxMap) SubscribeWith(ctx context.Context, sub rs.Subscriber) {
	p.source.SubscribeWith(ctx, newMapSubscriber(sub, p.mapper))
}

func newFluxMap(source Flux, mapper rs.Transformer) *fluxMap {
	return &fluxMap{
		source: source,
		mapper: mapper,
	}
}
