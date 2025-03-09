package subscribers

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

type BlockSubscriber[T any] struct {
	reactor.Item[T]
	doneChan chan struct{}
	ctxChan  chan struct{}
	done     int32
}

func NewBlockSubscriber[T any]() *BlockSubscriber[T] {
	return &BlockSubscriber[T]{
		doneChan: make(chan struct{}),
	}
}

func (b *BlockSubscriber[T]) Done() <-chan struct{} {
	return b.doneChan
}

func (b *BlockSubscriber[T]) OnComplete(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&b.done, 0, 1) {
		b.finish(ctx)
	}
}

func (b *BlockSubscriber[T]) OnError(ctx context.Context, err error) {
	if !atomic.CompareAndSwapInt32(&b.done, 0, 1) {
		// TODO: hooks
		//hooks.Global().OnErrorDrop(err)
		return
	}
	b.E = err
	b.finish(ctx)
}

func (b *BlockSubscriber[T]) OnNext(ctx context.Context, next T) {
	if atomic.LoadInt32(&b.done) != 0 || b.Valid || b.E != nil {
		// TODO: hooks
		//hooks.Global().OnNextDrop(next)
		return
	}
	b.Valid = true
	b.V = next
}

func (b *BlockSubscriber[T]) OnSubscribe(ctx context.Context, su reactor.Subscription) {
	// workaround: watch context
	if ctx != context.Background() && ctx != context.TODO() {
		ctxChan := make(chan struct{})
		b.ctxChan = ctxChan
		go func() {
			select {
			case <-ctx.Done():
				b.OnError(ctx, reactor.NewContextError(ctx.Err()))
			case <-ctxChan:
			}
		}()
	}
	su.Request(ctx, reactor.RequestInfinite)
}

func (b *BlockSubscriber[T]) finish(ctx context.Context) {
	if b.ctxChan != nil {
		close(b.ctxChan)
	}
	if b.doneChan != nil {
		close(b.doneChan)
	}

}
