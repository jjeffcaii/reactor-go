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
	actual    reactor.Subscriber
	values    []Any
	cursor    int32
	flags     uint8
	locker    sync.Mutex
	completed int32
}

func (ss *sliceSubscription) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	ss.locker.Lock()
	if ss.flags&flagFast != 0 {
		ss.locker.Unlock()
		return
	}
	if n < reactor.RequestInfinite {
		ss.locker.Unlock()
		ss.slowPath(n)
		return
	}
	ss.flags |= flagFast
	ss.locker.Unlock()
	ss.fastPath()
}

func (ss *sliceSubscription) Cancel() {
	ss.locker.Lock()
	ss.flags |= flagCancel
	ss.locker.Unlock()
}

func (ss *sliceSubscription) isCancelled() (cancelled bool) {
	ss.locker.Lock()
	cancelled = ss.flags&flagCancel != 0
	ss.locker.Unlock()
	return
}

func (ss *sliceSubscription) slowPath(n int) {
	for n > 0 {
		next := int(atomic.AddInt32(&ss.cursor, 1))
		if next > len(ss.values) {
			return
		}
		v := ss.values[next-1]
		if next == len(ss.values) {
			ss.actual.OnNext(v)
			ss.actual.OnComplete()
			return
		}
		n--
		ss.actual.OnNext(v)
	}
}

func (ss *sliceSubscription) fastPath() {
	i, l := int(ss.cursor), len(ss.values)
	if i >= l {
		return
	}
	for ; i < l; i++ {
		if ss.isCancelled() {
			return
		}
		v := ss.values[i]
		if v == nil {
			ss.actual.OnError(fmt.Errorf("the %dth slice element was null", i))
			return
		}
		ss.actual.OnNext(v)
	}
	if ss.isCancelled() {
		return
	}
	ss.actual.OnComplete()
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
	s.OnSubscribe(ctx, subscription)
}

func newSliceFlux(values []Any) *fluxSlice {
	return &fluxSlice{
		slice: values,
	}
}
