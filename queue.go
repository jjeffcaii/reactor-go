package rs

import (
	"context"
	"math"
	"sync/atomic"
)

type OnWaiting = func(ctx context.Context, q *Queue)

type Queue struct {
	tickets int32
	ch      chan interface{}
	pause   chan struct{}
}

func (p *Queue) Close() error {
	p.tickets = math.MaxInt32
	close(p.pause)
	close(p.ch)
	return nil
}

func (p *Queue) RequestN(n int32) {
	if atomic.CompareAndSwapInt32(&(p.tickets), 0, n) {
		p.pause <- struct{}{}
	} else {
		atomic.AddInt32(&(p.tickets), n)
	}
}

func (p *Queue) Add(v interface{}) (err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()
	p.ch <- v
	return
}

func (p *Queue) Poll(ctx context.Context, fn OnWaiting) (interface{}, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	case v, ok := <-p.ch:
		// tickets exhausted
		foo := atomic.LoadInt32(&(p.tickets))
		if foo == 0 {
			if fn != nil {
				fn(ctx, p)
			}
			<-p.pause
		}
		atomic.AddInt32(&(p.tickets), -1)
		return v, ok
	}
}

func NewQueue(size int) *Queue {
	return &Queue{
		tickets: 0,
		ch:      make(chan interface{}, size),
		pause:   make(chan struct{}, 1),
	}
}
