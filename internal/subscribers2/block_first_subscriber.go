package subscribers

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type BlockFirstSubscriber[T any] struct {
	su    reactor.Subscription
	done  chan struct{}
	vchan chan<- T
	echan chan<- error
}

func NewBlockFirstSubscriber[T any](done chan struct{}, vchan chan<- T, echan chan<- error) *BlockFirstSubscriber[T] {
	return &BlockFirstSubscriber[T]{
		done:  done,
		vchan: vchan,
		echan: echan,
	}
}

func (b *BlockFirstSubscriber[T]) OnComplete(ctx context.Context) {
	select {
	case <-b.done:
	default:
		internal.SafeCloseDone(b.done)
	}
}

func (b *BlockFirstSubscriber[T]) OnError(ctx context.Context, err error) {
	select {
	case <-b.done:
	default:
		if internal.SafeCloseDone(b.done) {
			b.echan <- err
		}
	}
}

func (b *BlockFirstSubscriber[T]) OnNext(ctx context.Context, next T) {
	select {
	case <-b.done:
	default:
		if internal.SafeCloseDone(b.done) {
			b.vchan <- next
			b.su.Cancel(ctx)
		}
	}
}

func (b *BlockFirstSubscriber[T]) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	select {
	case <-ctx.Done():
		b.OnError(ctx, reactor.NewContextError(ctx.Err()))
	default:
		b.su = subscription
		b.su.Request(ctx, 1)
	}
}
