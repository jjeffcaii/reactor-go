package subscribers

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

type blockSubscriber struct {
	done chan struct{}
	c    chan<- reactor.Item
}

func NewBlockSubscriber(done chan struct{}, c chan reactor.Item) reactor.Subscriber {
	return blockSubscriber{
		done: done,
		c:    c,
	}
}

func (b blockSubscriber) OnComplete() {
	select {
	case <-b.done:
	default:
		close(b.done)
	}
}

func (b blockSubscriber) OnError(err error) {
	select {
	case <-b.done:
	default:
		close(b.done)
		b.c <- reactor.Item{
			E: err,
		}
	}
}

func (b blockSubscriber) OnNext(any reactor.Any) {
	select {
	case <-b.done:
	default:
		b.c <- reactor.Item{
			V: any,
		}
	}
}

func (b blockSubscriber) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	// workaround: watch context
	if ctx != context.Background() && ctx != context.TODO() {
		go func() {
			select {
			case <-ctx.Done():
				b.OnError(reactor.ErrSubscribeCancelled)
			case <-b.done:
			}
		}()
	}
	subscription.Request(reactor.RequestInfinite)
}
