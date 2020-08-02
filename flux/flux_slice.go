package flux

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

const (
	flagCancel uint8 = 1 << iota
	flagFast
)

type sliceSubscription struct {
	actual reactor.Subscriber
	values []Any
	cursor int32
	flags  uint8
	locker sync.Mutex
}

func (p *sliceSubscription) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	p.locker.Lock()
	if p.flags&flagFast != 0 {
		p.locker.Unlock()
		return
	}
	if n < reactor.RequestInfinite {
		p.locker.Unlock()
		p.slowPath(n)
		return
	}
	p.flags |= flagFast
	p.locker.Unlock()
	p.fastPath()
}

func (p *sliceSubscription) Cancel() {
	p.locker.Lock()
	p.flags |= flagCancel
	p.locker.Unlock()
}

func (p *sliceSubscription) isCancelled() (cancelled bool) {
	p.locker.Lock()
	cancelled = p.flags&flagCancel != 0
	p.locker.Unlock()
	return
}

func (p *sliceSubscription) slowPath(n int) {
	for n > 0 {
		next := int(atomic.AddInt32(&(p.cursor), 1))
		if next > len(p.values) {
			return
		}
		v := p.values[next-1]
		if next == len(p.values) {
			p.actual.OnNext(v)
			p.actual.OnComplete()
			return
		}
		n--
		p.actual.OnNext(v)
	}
}

func (p *sliceSubscription) fastPath() {
	for i, l := int(p.cursor), len(p.values); i < l; i++ {
		if p.isCancelled() {
			return
		}
		v := p.values[i]
		if v == nil {
			p.actual.OnError(fmt.Errorf("the %dth slice element was null", i))
			return
		}
		p.actual.OnNext(v)
	}
	if p.isCancelled() {
		return
	}
	p.actual.OnComplete()
}

func newSliceSubscription(s reactor.Subscriber, values []Any) *sliceSubscription {
	return &sliceSubscription{
		actual: s,
		values: values,
	}
}

type fluxSlice struct {
	slice []Any
}

func (p *fluxSlice) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	if len(p.slice) < 1 {
		s.OnComplete()
		return
	}
	subscription := newSliceSubscription(s, p.slice)
	s.OnSubscribe(subscription)
}

func newSliceFlux(values []Any) *fluxSlice {
	return &fluxSlice{
		slice: values,
	}
}
