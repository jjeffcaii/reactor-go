package internal

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

type CoreSubscriber struct {
	reactor.Subscriber
	done chan struct{}
}

func (p *CoreSubscriber) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	if ctx != context.Background() && ctx != context.TODO() {
		go func() {
			select {
			case <-p.done:
			case <-ctx.Done():
				p.OnError(reactor.ErrSubscribeCancelled)
			}
		}()
	}

	select {
	case <-ctx.Done():
		p.Subscriber.OnError(reactor.ErrSubscribeCancelled)
	default:
		p.Subscriber.OnSubscribe(ctx, su)
	}
}

func (p *CoreSubscriber) OnComplete() {
	select {
	case <-p.done:
	default:
		close(p.done)
		p.Subscriber.OnComplete()
	}
}

func (p *CoreSubscriber) OnError(err error) {
	select {
	case <-p.done:
	default:
		close(p.done)
		p.Subscriber.OnError(err)
	}
}

func (p *CoreSubscriber) OnNext(v reactor.Any) {
	select {
	case <-p.done:
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
		done:       make(chan struct{}),
	}
}

func ExtractRawSubscriber(s reactor.Subscriber) reactor.Subscriber {
	if cs, ok := s.(*CoreSubscriber); ok {
		return cs.Subscriber
	}
	return s
}
