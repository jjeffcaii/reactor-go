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

func (p *CoreSubscriber) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	p.ctx = ctx
	select {
	case <-ctx.Done():
		p.Subscriber.OnError(reactor.ErrSubscribeCancelled)
	default:
		p.Subscriber.OnSubscribe(ctx, su)
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

func NewCoreSubscriber(actual reactor.Subscriber) *CoreSubscriber {
	if cs, ok := actual.(*CoreSubscriber); ok {
		return cs
	}
	return &CoreSubscriber{
		Subscriber: actual,
		ctx:        context.Background(),
	}
}

func ExtractRawSubscriber(s reactor.Subscriber) reactor.Subscriber {
	if cs, ok := s.(*CoreSubscriber); ok {
		return cs.Subscriber
	}
	return s
}
