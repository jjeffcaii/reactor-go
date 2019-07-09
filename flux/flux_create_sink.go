package flux

import (
	"io"
	"sync/atomic"

	rs "github.com/jjeffcaii/reactor-go"
)

type Sink interface {
	rs.Subscription
	Complete()
	Error(error)
	Next(interface{})
}

type queue interface {
	offer(interface{})
	poll() (interface{}, bool)
}

type simpleQueue struct {
	c chan interface{}
}

func (q simpleQueue) Close() (err error) {
	close(q.c)
	return
}

func (q simpleQueue) offer(v interface{}) {
	q.c <- v
}

func (q simpleQueue) poll() (v interface{}, ok bool) {
	select {
	case v, ok = <-q.c:
		return
	default:
		return
	}
}

type bufferedSink struct {
	s        rs.Subscriber
	q        queue
	n        int32
	draining int32
	done     bool
}

func (p *bufferedSink) Request(n int) {
	atomic.AddInt32(&(p.n), int32(n))
}

func (p *bufferedSink) Cancel() {
	p.dispose()
}

func (p *bufferedSink) Complete() {
	p.s.OnComplete()
	p.dispose()
}

func (p *bufferedSink) Error(err error) {
	p.s.OnError(err)
	p.dispose()
}

func (p *bufferedSink) Next(v interface{}) {
	if fs, ok := p.s.(filterSubscriber); ok {
		if !fs.predicate(v) {
			return
		}
	}
	p.q.offer(v)
	p.drain()
}

func (p *bufferedSink) drain() {
	if !atomic.CompareAndSwapInt32(&(p.draining), 0, 1) {
		return
	}
	defer atomic.StoreInt32(&(p.draining), 0)
	for atomic.AddInt32(&(p.n), -1) > -1 {
		if p.done {
			return
		}
		v, ok := p.q.poll()
		if !ok {
			atomic.AddInt32(&(p.n), 1)
			break
		}
		p.s.OnNext(v)
	}
	atomic.CompareAndSwapInt32(&(p.n), -1, 0)
}

func (p *bufferedSink) dispose() {
	p.done = true
	if closer, ok := p.q.(io.Closer); ok {
		_ = closer.Close()
	}
}

func newBufferedSink(s rs.Subscriber, cap int) Sink {
	return &bufferedSink{
		s: s,
		q: simpleQueue{
			c: make(chan interface{}, cap),
		},
	}
}
