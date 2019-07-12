package internal

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

type CoreSubscriber struct {
	rs.Subscriber
	ctx context.Context
}

func (p *CoreSubscriber) OnSubscribe(su rs.Subscription) {
	select {
	case <-p.ctx.Done():
		p.Subscriber.OnError(rs.ErrSubscribeCancelled)
	default:
		p.Subscriber.OnSubscribe(su)
	}
}

func (p *CoreSubscriber) OnNext(v interface{}) {
	select {
	case <-p.ctx.Done():
		hooks.Global().OnNextDrop(v)
		p.Subscriber.OnError(rs.ErrSubscribeCancelled)
	default:
		p.Subscriber.OnNext(v)
	}
}

func NewCoreSubscriber(ctx context.Context, actual rs.Subscriber) *CoreSubscriber {
	if cs, ok := actual.(*CoreSubscriber); ok {
		return cs
	}
	return &CoreSubscriber{
		Subscriber: actual,
		ctx:        ctx,
	}
}

func ExtractRawSubscriber(s rs.Subscriber) rs.Subscriber {
	if cs, ok := s.(*CoreSubscriber); ok {
		return cs.Subscriber
	}
	return s
}
