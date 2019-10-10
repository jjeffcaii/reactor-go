package flux

import (
	"sync"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"go.uber.org/atomic"
)

type bufferedSink struct {
	s        rs.Subscriber
	q        queue
	n        *atomic.Int32
	draining *atomic.Bool
	stat     *atomic.Int32
	cond     *sync.Cond
}

func (p *bufferedSink) Request(n int) {
	p.n.Add(int32(n))
	p.drain()
}

func (p *bufferedSink) Cancel() {
	if !p.stat.CAS(0, statCancel) {
		return
	}
	// TODO: support cancel
	p.dispose()
}

func (p *bufferedSink) Complete() {
	p.cond.L.Lock()
	for p.draining.Load() || p.q.size() > 0 {
		p.cond.Wait()
	}
	p.cond.L.Unlock()
	if p.stat.CAS(0, statComplete) {
		p.s.OnComplete()
		p.dispose()
	}
}

func (p *bufferedSink) Error(err error) {
	if p.stat.CAS(0, statError) {
		hooks.Global().OnErrorDrop(err)
		return
	}
	p.s.OnError(err)
	p.dispose()
}

func (p *bufferedSink) Next(v interface{}) {
	if p.stat.Load() != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	p.q.offer(v)
	p.drain()
}

func (p *bufferedSink) drain() {
	if !p.draining.CAS(false, true) {
		return
	}
	defer func() {
		p.cond.L.Lock()
		p.draining.CAS(true, false)
		p.cond.Broadcast()
		p.cond.L.Unlock()
	}()
	for p.n.Add(-1) > -1 {
		if p.stat.Load() != 0 {
			return
		}
		v, ok := p.q.poll()
		if !ok {
			p.n.Inc()
			break
		}
		p.s.OnNext(v)
	}
	p.n.CAS(-1, 0)
}

func (p *bufferedSink) dispose() {
	_ = p.q.Close()
}

func newBufferedSink(s rs.Subscriber, cap int) *bufferedSink {
	return &bufferedSink{
		s:        s,
		q:        newQueue(cap),
		n:        atomic.NewInt32(0),
		draining: atomic.NewBool(false),
		stat:     atomic.NewInt32(0),
		cond:     sync.NewCond(&sync.Mutex{}),
	}
}
