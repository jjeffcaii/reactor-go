package flux

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

type mapSubscriber struct {
	source reactor.Subscriber
	mapper reactor.Transformer
}

func (s mapSubscriber) OnComplete() {
	s.source.OnComplete()
}

func (s mapSubscriber) OnError(err error) {
	s.source.OnError(err)
}

func (s mapSubscriber) OnNext(i Any) {
	if o, err := s.mapper(i); err != nil {
		s.source.OnError(err)
	} else {
		s.source.OnNext(o)
	}
}

func (s mapSubscriber) OnSubscribe(subscription reactor.Subscription) {
	s.source.OnSubscribe(subscription)
}

func newMapSubscriber(s reactor.Subscriber, mapper reactor.Transformer) reactor.Subscriber {
	return mapSubscriber{
		source: s,
		mapper: mapper,
	}
}

type fluxMap struct {
	source reactor.RawPublisher
	mapper reactor.Transformer
}

func (p *fluxMap) SubscribeWith(ctx context.Context, sub reactor.Subscriber) {
	p.source.SubscribeWith(ctx, newMapSubscriber(sub, p.mapper))
}

func newFluxMap(source reactor.RawPublisher, mapper reactor.Transformer) *fluxMap {
	return &fluxMap{
		source: source,
		mapper: mapper,
	}
}
