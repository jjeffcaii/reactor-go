package flux

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
)

var errDuplicatedSubscriber = errors.New("UnicastProcessor allows only a single Subscriber")

type rawProcessor interface {
	rs.RawPublisher
	Sink
}

type unicastProcessor struct {
	q          queue
	actual     rs.Subscriber
	locker     sync.Mutex
	stat       int32
	n          int32
	draining   int32
	subscribed chan struct{}
}

func (p *unicastProcessor) Request(n int) {
	atomic.AddInt32(&(p.n), int32(n))
}

func (p *unicastProcessor) Cancel() {
	p.dispose(statCancel)
}

func (p *unicastProcessor) OnComplete() {
	println("....begin complete")
	if p == nil {
		log.Println("A")
	}
	if p.dispose(statComplete) {
		p.drain()
		if p == nil {
			println("C")
		}
		if p.actual == nil {
			println("B")
		}
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

func (p *unicastProcessor) OnNext(v interface{}) {
	p.q.offer(v)
	p.drain()
}

func (p *unicastProcessor) OnSubscribe(su rs.Subscription) {
	p.actual.OnSubscribe(su)
}

func (p *unicastProcessor) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	p.locker.Lock()
	if p.actual != nil {
		p.locker.Unlock()
		panic(errDuplicatedSubscriber)
	}
	raw := internal.ExtractRawSubscriber(s)
	p.actual = raw
	p.locker.Unlock()
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

func (p *unicastProcessor) Next(v interface{}) {
	<-p.subscribed
	p.OnNext(v)
}

func (p *unicastProcessor) drain() {
	if !atomic.CompareAndSwapInt32(&(p.draining), 0, 1) {
		return
	}
	defer atomic.StoreInt32(&(p.draining), 0)
	for atomic.AddInt32(&(p.n), -1) > -1 {
		if atomic.LoadInt32(&(p.stat)) != 0 {
			return
		}
		v, ok := p.q.poll()
		if !ok {
			atomic.AddInt32(&(p.n), 1)
			break
		}
		p.actual.OnNext(v)
	}
	atomic.CompareAndSwapInt32(&(p.n), -1, 0)
}

func (p *unicastProcessor) dispose(stat int32) bool {
	if !atomic.CompareAndSwapInt32(&(p.stat), 0, stat) {
		return false
	}
	_ = p.q.Close()
	return true
}

func newUnicastProcessor(cap int) *unicastProcessor {
	return &unicastProcessor{
		q:          newQueue(cap),
		subscribed: make(chan struct{}),
	}
}
