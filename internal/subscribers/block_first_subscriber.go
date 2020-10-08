package subscribers

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type blockFirstSubscriber struct {
	su    reactor.Subscription
	done  chan struct{}
	vchan chan<- reactor.Any
	echan chan<- error
}

func NewBlockFirstSubscriber(done chan struct{}, vchan chan<- reactor.Any, echan chan<- error) reactor.Subscriber {
	return &blockFirstSubscriber{
		done:  done,
		vchan: vchan,
		echan: echan,
	}
}

func (b *blockFirstSubscriber) OnComplete() {
	select {
	case <-b.done:
	default:
		internal.SafeCloseDone(b.done)
	}
}

func (b *blockFirstSubscriber) OnError(err error) {
	select {
	case <-b.done:
	default:
		if internal.SafeCloseDone(b.done) {
			b.echan <- err
		}
	}
}

func (b *blockFirstSubscriber) OnNext(any reactor.Any) {
	select {
	case <-b.done:
	default:
		if internal.SafeCloseDone(b.done) {
			b.vchan <- any
			b.su.Cancel()
		}
	}
}

func (b *blockFirstSubscriber) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	select {
	case <-ctx.Done():
		b.OnError(reactor.ErrSubscribeCancelled)
	default:
		b.su = subscription
		b.su.Request(1)
	}
}
