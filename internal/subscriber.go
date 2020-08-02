package internal

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

type CoreSubscriber struct {
	reactor.Subscriber
	ctx context.Context
}

func (p *CoreSubscriber) OnSubscribe(su reactor.Subscription) {
	select {
	case <-p.ctx.Done():
		p.Subscriber.OnError(reactor.ErrSubscribeCancelled)
	default:
		p.Subscriber.OnSubscribe(su)
	}
}

func (p *CoreSubscriber) OnNext(v reactor.Any) {
	select {
	case <-p.ctx.Done():
		hooks.Global().OnNextDrop(v)
		p.Subscriber.OnError(reactor.ErrSubscribeCancelled)
	default:
		p.Subscriber.OnNext(v)
	}
}

func NewCoreSubscriber(ctx context.Context, actual reactor.Subscriber) *CoreSubscriber {
	if cs, ok := actual.(*CoreSubscriber); ok {
		return cs
	}
	return &CoreSubscriber{
		Subscriber: actual,
		ctx:        ctx,
	}
}

func ExtractRawSubscriber(s reactor.Subscriber) reactor.Subscriber {
	if cs, ok := s.(*CoreSubscriber); ok {
		return cs.Subscriber
	}
	return s
}
