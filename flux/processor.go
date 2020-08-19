package flux

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
)

type rawProcessor interface {
	reactor.RawPublisher
	Sink
}

type unicastProcessor struct {
	q          queue
	actual     reactor.Subscriber
	stat       int32
	requests   int32
	draining   int32
	subscribed chan struct{}
	lock       sync.RWMutex
	cond       sync.Cond
}

func (up *unicastProcessor) Request(n int) {
	atomic.AddInt32(&up.requests, int32(n))
}

func (up *unicastProcessor) Cancel() {
	up.dispose(statCancel)
}

func (up *unicastProcessor) OnComplete() {
	up.cond.L.Lock()
	for atomic.LoadInt32(&up.draining) != 0 {
		up.cond.Wait()
	}
	up.cond.L.Unlock()
	if up.dispose(statComplete) {
		up.drain()
		up.actual.OnComplete()
	}
}

func (up *unicastProcessor) OnError(e error) {
	if up.dispose(statError) {
		up.actual.OnError(e)
	} else {
		hooks.Global().OnErrorDrop(e)
	}
}

func (up *unicastProcessor) OnNext(v Any) {
	if atomic.LoadInt32(&up.stat) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	up.q.offer(v)
	up.drain()
}

func (up *unicastProcessor) OnSubscribe(su reactor.Subscription) {
	up.actual.OnSubscribe(su)
}

func (up *unicastProcessor) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	up.lock.RLock()
	conflict := up.actual != nil
	up.lock.RUnlock()
	if conflict {
		panic(errSubscribeOnce)
	}

	up.cond.L.Lock()
	raw := internal.ExtractRawSubscriber(s)
	up.actual = raw
	up.cond.L.Unlock()

	defer close(up.subscribed)
	raw.OnSubscribe(up)
}

func (up *unicastProcessor) Complete() {
	select {
	case <-up.subscribed:
		up.OnComplete()
	default:
	}
}

func (up *unicastProcessor) Error(e error) {
	select {
	case <-up.subscribed:
		up.OnError(e)
	default:
	}
}

func (up *unicastProcessor) Next(v Any) {
	select {
	case <-up.subscribed:
		up.OnNext(v)
	default:
	}
}

func (up *unicastProcessor) drain() {
	if !atomic.CompareAndSwapInt32(&up.draining, 0, 1) {
		return
	}
	defer func() {
		up.cond.L.Lock()
		atomic.StoreInt32(&up.draining, 0)
		up.cond.Broadcast()
		up.cond.L.Unlock()
	}()
	for atomic.AddInt32(&up.requests, -1) > -1 {
		if atomic.LoadInt32(&up.stat) != 0 {
			return
		}
		v, ok := up.q.poll()
		if !ok {
			atomic.AddInt32(&up.requests, 1)
			break
		}
		up.actual.OnNext(v)
	}
	atomic.CompareAndSwapInt32(&up.requests, -1, 0)
}

func (up *unicastProcessor) dispose(stat int32) (ok bool) {
	ok = atomic.CompareAndSwapInt32(&up.stat, 0, stat)
	if ok {
		_ = up.q.Close()
	}
	return
}

func newUnicastProcessor(cap int) *unicastProcessor {
	u := &unicastProcessor{
		q:          newQueue(cap),
		subscribed: make(chan struct{}),
	}
	u.cond.L = &u.lock
	return u
}
