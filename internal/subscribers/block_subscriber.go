package subscribers

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
)

type blockSubscriber struct {
	done  chan struct{}
	vchan chan<- reactor.Any
	echan chan<- error
}

func NewBlockSubscriber(done chan struct{}, vchan chan reactor.Any, echan chan error) reactor.Subscriber {
	return blockSubscriber{
		done:  done,
		vchan: vchan,
		echan: echan,
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
		b.echan <- err
	}
}

func (b blockSubscriber) OnNext(any reactor.Any) {
	select {
	case <-b.done:
	default:
		b.vchan <- any
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
