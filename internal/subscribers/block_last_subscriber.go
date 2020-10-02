package subscribers

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type blockLastSubscriber struct {
	done     chan struct{}
	callback func(reactor.Any, error)
}

func NewBlockLastSubscriber(done chan struct{}, callback func(reactor.Any, error)) reactor.Subscriber {
	return blockLastSubscriber{
		done:     done,
		callback: callback,
	}
}

func (b blockLastSubscriber) OnComplete() {
	internal.SafeCloseDone(b.done)
}

func (b blockLastSubscriber) OnError(err error) {
	select {
	case <-b.done:
	default:
		if internal.SafeCloseDone(b.done) {
			b.callback(nil, err)
		}
	}
}

func (b blockLastSubscriber) OnNext(any reactor.Any) {
	select {
	case <-b.done:
	default:
		b.callback(any, nil)
	}
}

func (b blockLastSubscriber) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	select {
	case <-ctx.Done():
	default:
		subscription.Request(reactor.RequestInfinite)
	}
}
