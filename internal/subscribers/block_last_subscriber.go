package subscribers

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
)

var _blockLastSubscriberPool = sync.Pool{
	New: func() interface{} {
		return &BlockLastSubscriber{
			doneNotify: make(chan struct{}, 1),
			prev:       new(internal.Item),
		}
	},
}

var _ reactor.Subscriber = (*BlockLastSubscriber)(nil)

type BlockLastSubscriber struct {
	doneNotify chan struct{}
	prev       *internal.Item
	done       int32
}

func BorrowBlockLastSubscriber() *BlockLastSubscriber {
	return _blockLastSubscriberPool.Get().(*BlockLastSubscriber)
}

func ReturnBlockLastSubscriber(s *BlockLastSubscriber) {
	s.prev.V = nil
	s.prev.E = nil
	atomic.StoreInt32(&s.done, 0)
	_blockLastSubscriberPool.Put(s)
}

func (b *BlockLastSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&b.done, 0, 1) {
		b.doneNotify <- struct{}{}
	}
}

func (b *BlockLastSubscriber) OnError(err error) {
	if !atomic.CompareAndSwapInt32(&b.done, 0, 1) {
		hooks.Global().OnErrorDrop(err)
		return
	}
	if b.prev.V != nil {
		hooks.Global().OnNextDrop(b.prev.V)
	}
	b.prev.V = nil
	b.prev.E = err
	b.doneNotify <- struct{}{}
}

func (b *BlockLastSubscriber) OnNext(value reactor.Any) {
	if atomic.LoadInt32(&b.done) != 0 {
		hooks.Global().OnNextDrop(value)
		return
	}
	if b.prev.V != nil {
		hooks.Global().OnNextDrop(value)
	}
	b.prev.V = value
}

func (b *BlockLastSubscriber) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	select {
	case <-ctx.Done():
		b.OnError(reactor.ErrSubscribeCancelled)
	default:
		subscription.Request(reactor.RequestInfinite)
	}
}

func (b *BlockLastSubscriber) Block() (value reactor.Any, err error) {
	<-b.doneNotify
	value = b.prev.V
	err = b.prev.E
	return
}
