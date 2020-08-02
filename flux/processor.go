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
	cond       *sync.Cond
}

func (p *unicastProcessor) Request(n int) {
	atomic.AddInt32(&p.requests, int32(n))
}

func (p *unicastProcessor) Cancel() {
	p.dispose(statCancel)
}

func (p *unicastProcessor) OnComplete() {
	p.cond.L.Lock()
	for atomic.LoadInt32(&p.draining) != 0 {
		p.cond.Wait()
	}
	p.cond.L.Unlock()
	if p.dispose(statComplete) {
		p.drain()
		p.actual.OnComplete()
	}
}

func (p *unicastProcessor) OnError(e error) {
	if p.dispose(statError) {
		p.actual.OnError(e)
	} else {
		hooks.Global().OnErrorDrop(e)
	}
}

func (p *unicastProcessor) OnNext(v Any) {
	if atomic.LoadInt32(&p.stat) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	p.q.offer(v)
	p.drain()
}

func (p *unicastProcessor) OnSubscribe(su reactor.Subscription) {
	p.actual.OnSubscribe(su)
}

func (p *unicastProcessor) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	p.cond.L.Lock()
	if p.actual != nil {
		p.cond.L.Unlock()
		panic(errSubscribeOnce)
	}
	raw := internal.ExtractRawSubscriber(s)
	p.actual = raw
	p.cond.L.Unlock()
	raw.OnSubscribe(p)
	close(p.subscribed)
}

func (p *unicastProcessor) Complete() {
	<-p.subscribed
	p.OnComplete()
}

func (p *unicastProcessor) Error(e error) {
	<-p.subscribed
	p.OnError(e)
}

func (p *unicastProcessor) Next(v Any) {
	<-p.subscribed
	p.OnNext(v)
}

func (p *unicastProcessor) drain() {
	if !atomic.CompareAndSwapInt32(&p.draining, 0, 1) {
		return
	}
	defer func() {
		p.cond.L.Lock()
		atomic.StoreInt32(&p.draining, 0)
		p.cond.Broadcast()
		p.cond.L.Unlock()
	}()
	for atomic.AddInt32(&p.requests, -1) > -1 {
		if atomic.LoadInt32(&p.stat) != 0 {
			return
		}
		v, ok := p.q.poll()
		if !ok {
			atomic.AddInt32(&p.requests, 1)
			break
		}
		p.actual.OnNext(v)
	}
	atomic.CompareAndSwapInt32(&p.requests, -1, 0)
}

func (p *unicastProcessor) dispose(stat int32) (ok bool) {
	ok = atomic.CompareAndSwapInt32(&p.stat, 0, stat)
	if ok {
		_ = p.q.Close()
	}
	return
}

func newUnicastProcessor(cap int) *unicastProcessor {
	return &unicastProcessor{
		q:          newQueue(cap),
		subscribed: make(chan struct{}),
		cond:       sync.NewCond(&sync.Mutex{}),
	}
}
