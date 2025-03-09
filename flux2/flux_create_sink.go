package flux

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

type bufferedSink[T any] struct {
	s        reactor.Subscriber[T]
	q        queue[T]
	n        int32
	draining int32
	stat     int32
	lock     sync.Mutex
	cond     sync.Cond
}

func newBufferedSink[T any](s reactor.Subscriber[T], cap int) *bufferedSink[T] {
	b := &bufferedSink[T]{
		s: s,
		q: newQueue[T](cap),
	}
	b.cond.L = &b.lock
	return b
}

func (p *bufferedSink[T]) Request(ctx context.Context, n int) {
	atomic.AddInt32(&p.n, int32(n))
	p.drain(ctx)
}

func (p *bufferedSink[T]) Cancel(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statCancel) {
		return
	}
	// TODO: support cancel
	p.dispose()
}

func (p *bufferedSink[T]) Complete(ctx context.Context) {
	p.cond.L.Lock()
	for atomic.LoadInt32(&p.draining) == 1 || p.q.size() > 0 {
		p.cond.Wait()
	}
	p.cond.L.Unlock()
	if atomic.CompareAndSwapInt32(&p.stat, 0, statComplete) {
		p.s.OnComplete(ctx)
		p.dispose()
	}
}

func (p *bufferedSink[T]) Error(ctx context.Context, err error) {
	if atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		p.s.OnError(ctx, err)
		p.dispose()
		return
	}
	// TODO: hooks
	//hooks.Global().OnErrorDrop(err)
}

func (p *bufferedSink[T]) Next(ctx context.Context, next T) {
	if atomic.LoadInt32(&p.stat) != 0 {
		// TODO: hooks
		//hooks.Global().OnNextDrop(v)
		return
	}
	p.q.offer(next)
	p.drain(ctx)
}

func (p *bufferedSink[T]) drain(ctx context.Context) {
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
		p.s.OnNext(ctx, v)
	}
	atomic.CompareAndSwapInt32(&p.n, -1, 0)
}

func (p *bufferedSink[T]) dispose() {
	_ = p.q.Close()
}
