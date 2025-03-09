package subscribers

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

type BlockLastSubscriber[T any] struct {
	doneNotify chan struct{}
	prev       reactor.Item[T]
	done       int32
}

func NewBlockLastSubscriber[T any]() *BlockLastSubscriber[T] {
	return &BlockLastSubscriber[T]{
		doneNotify: make(chan struct{}),
	}
}

func (b *BlockLastSubscriber[T]) OnComplete(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&b.done, 0, 1) {
		close(b.doneNotify)
	}
}

func (b *BlockLastSubscriber[T]) OnError(ctx context.Context, err error) {
	if !atomic.CompareAndSwapInt32(&b.done, 0, 1) {
		// TODO: hooks
		//hooks.Global().OnErrorDrop(err)
		return
	}
	if b.prev.Valid {
		// TODO: hooks
		//hooks.Global().OnNextDrop(b.prev.V)
	}

	var zero T

	b.prev.Valid = false
	b.prev.V = zero
	b.prev.E = err

	close(b.doneNotify)
}

func (b *BlockLastSubscriber[T]) OnNext(ctx context.Context, value T) {
	if atomic.LoadInt32(&b.done) != 0 {
		// TODO: hooks
		//hooks.Global().OnNextDrop(value)
		return
	}
	if b.prev.Valid {
		// TODO: hooks
		//hooks.Global().OnNextDrop(value)
	}

	b.prev.Valid = true
	b.prev.V = value
}

func (b *BlockLastSubscriber[T]) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	select {
	case <-ctx.Done():
		b.OnError(ctx, reactor.NewContextError(ctx.Err()))
	default:
		subscription.Request(ctx, reactor.RequestInfinite)
	}
}

func (b *BlockLastSubscriber[T]) Block() (value T, ok bool, err error) {
	<-b.doneNotify

	if b.prev.Valid {
		ok = true
		value = b.prev.V
	} else {
		err = b.prev.E
	}

	return
}
