package flux

import (
	"context"
	"sync"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
	"go.uber.org/atomic"
)

type rawProcessor interface {
	rs.RawPublisher
	Sink
}

type unicastProcessor struct {
	q          queue
	actual     rs.Subscriber
	stat       *atomic.Int32
	requests   *atomic.Int32
	draining   *atomic.Bool
	subscribed chan struct{}
	cond       *sync.Cond
}

func (p *unicastProcessor) Request(n int) {
	p.requests.Add(int32(n))
}

func (p *unicastProcessor) Cancel() {
	p.dispose(statCancel)
}

func (p *unicastProcessor) OnComplete() {
	p.cond.L.Lock()
	for p.draining.Load() {
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

func (p *unicastProcessor) OnNext(v interface{}) {
	if p.stat.Load() != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	p.q.offer(v)
	p.drain()
}

func (p *unicastProcessor) OnSubscribe(su rs.Subscription) {
	p.actual.OnSubscribe(su)
}

func (p *unicastProcessor) SubscribeWith(ctx context.Context, s rs.Subscriber) {
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

func (p *unicastProcessor) Next(v interface{}) {
	<-p.subscribed
	p.OnNext(v)
}

func (p *unicastProcessor) drain() {
	if !p.draining.CAS(false, true) {
		return
	}
	defer func() {
		p.cond.L.Lock()
		p.draining.Store(false)
		p.cond.Broadcast()
		p.cond.L.Unlock()
	}()
	for p.requests.Dec() > -1 {
		if p.stat.Load() != 0 {
			return
		}
		v, ok := p.q.poll()
		if !ok {
			p.requests.Inc()
			break
		}
		p.actual.OnNext(v)
	}
	p.requests.CAS(-1, 0)
}

func (p *unicastProcessor) dispose(stat int32) (ok bool) {
	ok = p.stat.CAS(0, stat)
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
		stat:       atomic.NewInt32(0),
		requests:   atomic.NewInt32(0),
		draining:   atomic.NewBool(false),
	}
}
