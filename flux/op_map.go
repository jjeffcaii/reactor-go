package flux

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
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

func (s mapSubscriber) Raw() rs.RawSubscriber {
	panic("not implement")
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

func (p fluxMap) Filter(filter rs.Predicate) Flux {
	return newFluxFilter(p, filter)
}

func (p fluxMap) Map(tf rs.Transformer) Flux {
	return newFluxMap(p, tf)
}

func (p fluxMap) SubscribeOn(sc scheduler.Scheduler) Flux {
	return newFluxSubscribeOn(p, sc)
}

func (p fluxMap) Subscribe(ctx context.Context, sub rs.Subscriber) rs.Disposable {
	p.source.Subscribe(ctx, newMapSubscriber(sub, p.mapper))
	return nil
}

func newFluxMap(source Flux, mapper rs.Transformer) Flux {
	return fluxMap{
		source: source,
		mapper: mapper,
	}
}
