package flux

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"
)

const (
	RequestInfinite = math.MaxInt32

	DefaultQueueSize = 16
)

var errIllegalCap = errors.New("cap must greater than zero")

type Queue struct {
	elements   chan interface{}
	cond       *sync.Cond
	tickets    int32
	onRequestN func(int32)
	done       chan struct{}
}

func (p *Queue) Close() (err error) {
	p.cond.Broadcast()
	close(p.elements)
	return
}

func (p *Queue) HandleRequest(handler func(n int32)) {
	p.onRequestN = handler
}

func (p *Queue) SetTickets(n int32) {
	atomic.StoreInt32(&(p.tickets), n)
}

func (p *Queue) Tickets() (n int32) {
	n = atomic.LoadInt32(&(p.tickets))
	if n < 0 {
		n = 0
	}
	return
}

func (p *Queue) Push(item interface{}) (err error) {
	defer func() {
		err, _ = recover().(error)
	}()
	p.elements <- item
	return
}

func (p *Queue) Request(n int32) {
	if n < 1 {
		return
	}
	p.cond.L.Lock()
	if atomic.LoadInt32(&(p.tickets)) < 1 {
		atomic.StoreInt32(&(p.tickets), n)
		p.cond.Signal()
	} else {
		atomic.StoreInt32(&(p.tickets), n)
	}
	if p.onRequestN != nil {
		p.onRequestN(n)
	}
	p.cond.L.Unlock()
}

func (p *Queue) Poll() (item interface{}, ok bool) {
	select {
	case <-p.done:
		return
	default:
		p.cond.L.Lock()
		if atomic.LoadInt32(&(p.tickets)) == RequestInfinite {
			item, ok = <-p.elements
			if !ok {
				close(p.done)
			}
			p.cond.L.Unlock()
			return
		}
		for atomic.AddInt32(&(p.tickets), -1) < 0 {
			p.cond.Wait()
		}
		item, ok = <-p.elements
		if !ok {
			close(p.done)
		}
		p.cond.L.Unlock()
	}
	return
}

func NewQueue(cap int, tickets int32) *Queue {
	if cap < 1 {
		panic(errIllegalCap)
	}
	return &Queue{
		cond:     sync.NewCond(&sync.Mutex{}),
		tickets:  tickets,
		elements: make(chan interface{}, cap),
		done:     make(chan struct{}),
	}
}
