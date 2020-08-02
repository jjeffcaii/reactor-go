package flux

import (
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

type bufferedSink struct {
	s        reactor.Subscriber
	q        queue
	n        int32
	draining int32
	stat     int32
	cond     *sync.Cond
}

func (p *bufferedSink) Request(n int) {
	atomic.AddInt32(&p.n, int32(n))
	p.drain()
}

func (p *bufferedSink) Cancel() {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statCancel) {
		return
	}
	// TODO: support cancel
	p.dispose()
}

func (p *bufferedSink) Complete() {
	p.cond.L.Lock()
	for atomic.LoadInt32(&p.draining) == 1 || p.q.size() > 0 {
		p.cond.Wait()
	}
	p.cond.L.Unlock()
	if atomic.CompareAndSwapInt32(&p.stat, 0, statComplete) {
		p.s.OnComplete()
		p.dispose()
	}
}

func (p *bufferedSink) Error(err error) {
	if atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		p.s.OnError(err)
		p.dispose()
		return
	}
	hooks.Global().OnErrorDrop(err)
}

func (p *bufferedSink) Next(v Any) {
	if atomic.LoadInt32(&p.stat) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	p.q.offer(v)
	p.drain()
}

func (p *bufferedSink) drain() {
	if !atomic.CompareAndSwapInt32(&p.draining, 0, 1) {
		return
	}
	defer func() {
		p.cond.L.Lock()
		atomic.CompareAndSwapInt32(&p.draining, 1, 0)
		p.cond.Broadcast()
		p.cond.L.Unlock()
	}()
	for atomic.AddInt32(&p.n, -1) > -1 {
		if atomic.LoadInt32(&p.stat) != 0 {
			return
		}
		v, ok := p.q.poll()
		if !ok {
			atomic.AddInt32(&p.n, 1)
			break
		}
		p.s.OnNext(v)
	}
	atomic.CompareAndSwapInt32(&p.n, -1, 0)
}

func (p *bufferedSink) dispose() {
	_ = p.q.Close()
}

func newBufferedSink(s reactor.Subscriber, cap int) *bufferedSink {
	return &bufferedSink{
		s:    s,
		q:    newQueue(cap),
		cond: sync.NewCond(&sync.Mutex{}),
	}
}
