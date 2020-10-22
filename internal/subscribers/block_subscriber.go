package subscribers

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type blockSubscriber struct {
	done chan struct{}
	c    chan<- internal.Item
}

func NewBlockSubscriber(done chan struct{}, c chan internal.Item) reactor.Subscriber {
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
		b.c <- internal.Item{
			E: err,
		}
	}
}

func (b blockSubscriber) OnNext(any reactor.Any) {
	select {
	case <-b.done:
	default:
		b.c <- internal.Item{
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
