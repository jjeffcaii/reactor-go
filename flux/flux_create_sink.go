package flux

import (
  "sync"
  "sync/atomic"

  rs "github.com/jjeffcaii/reactor-go"
  "github.com/jjeffcaii/reactor-go/hooks"
)

type bufferedSink struct {
	s        rs.Subscriber
	q        queue
	n        int32
	draining int32
	done     bool
	cond     *sync.Cond
}

func (p *bufferedSink) Request(n int) {
	atomic.AddInt32(&(p.n), int32(n))
	p.drain()
}

func (p *bufferedSink) Cancel() {
	// TODO: support cancel
	p.dispose()
}

func (p *bufferedSink) Complete() {
	p.cond.L.Lock()
	for atomic.LoadInt32(&(p.draining)) == 1 || p.q.size() > 0 {
		p.cond.Wait()
	}
	p.cond.L.Unlock()
	p.s.OnComplete()
	p.dispose()
}

func (p *bufferedSink) Error(err error) {
	if p.done {
		hooks.Global().OnErrorDrop(err)
		return
	}
	p.s.OnError(err)
	p.dispose()
}

func (p *bufferedSink) Next(v interface{}) {
	if p.done {
		hooks.Global().OnNextDrop(v)
		return
	}
	p.q.offer(v)
	p.drain()
}

func (p *bufferedSink) drain() {
	if !atomic.CompareAndSwapInt32(&(p.draining), 0, 1) {
		return
	}
	defer func() {
		p.cond.L.Lock()
		atomic.StoreInt32(&(p.draining), 0)
		p.cond.Broadcast()
		p.cond.L.Unlock()
	}()
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
	_ = p.q.Close()
}

func newBufferedSink(s rs.Subscriber, cap int) *bufferedSink {
	return &bufferedSink{
		s:    s,
		q:    newQueue(cap),
		cond: sync.NewCond(&sync.Mutex{}),
	}
}
