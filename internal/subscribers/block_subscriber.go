package subscribers

import (
	"context"
	"math"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

var globalBlockSubscriberPool blockSubscriberPool

func BorrowBlockSubscriber() *BlockSubscriber {
	return globalBlockSubscriberPool.get()
}

func ReturnBlockSubscriber(s *BlockSubscriber) {
	globalBlockSubscriberPool.put(s)
}

type blockSubscriberPool struct {
	inner sync.Pool
}

func (bp *blockSubscriberPool) get() *BlockSubscriber {
	if exist, _ := bp.inner.Get().(*BlockSubscriber); exist != nil {
		atomic.StoreInt32(&exist.done, 0)
		return exist
	}
	return &BlockSubscriber{
		doneChan: make(chan struct{}, 1),
	}
}

func (bp *blockSubscriberPool) put(s *BlockSubscriber) {
	if s == nil {
		return
	}
	s.Reset()
	bp.inner.Put(s)
}

type BlockSubscriber struct {
	reactor.Item
	doneChan chan struct{}
	ctxChan  chan struct{}
	done     int32
}

func (b *BlockSubscriber) Reset() {
	b.V = nil
	b.E = nil
	b.ctxChan = nil
	atomic.StoreInt32(&b.done, math.MinInt32)
}

func (b *BlockSubscriber) Done() <-chan struct{} {
	return b.doneChan
}

func (b *BlockSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&b.done, 0, 1) {
		b.finish()
	}
}

func (b *BlockSubscriber) OnError(err error) {
	if !atomic.CompareAndSwapInt32(&b.done, 0, 1) {
		hooks.Global().OnErrorDrop(err)
		return
	}
	b.E = err
	b.finish()
}

func (b *BlockSubscriber) finish() {
	if b.ctxChan != nil {
		close(b.ctxChan)
	}
	b.doneChan <- struct{}{}
}

func (b *BlockSubscriber) OnNext(any reactor.Any) {
	if atomic.LoadInt32(&b.done) != 0 || b.V != nil || b.E != nil {
		hooks.Global().OnNextDrop(any)
		return
	}
	b.V = any
}

func (b *BlockSubscriber) OnSubscribe(ctx context.Context, subscription reactor.Subscription) {
	// workaround: watch context
	if ctx != context.Background() && ctx != context.TODO() {
		ctxChan := make(chan struct{})
		b.ctxChan = ctxChan
		go func() {
			select {
			case <-ctx.Done():
				b.OnError(reactor.NewContextError(ctx.Err()))
			case <-ctxChan:
			}
		}()
	}
	subscription.Request(reactor.RequestInfinite)
}
