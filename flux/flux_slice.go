package flux

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

type sliceSubscription struct {
	actual    rs.Subscriber
	values    []interface{}
	cursor    int32
	cancelled int32
}

func (p *sliceSubscription) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
	}
	if n >= rs.RequestInfinite {
		p.fastPath()
	} else {
		p.slowPath(n)
	}
}

func (p *sliceSubscription) Cancel() {
	atomic.StoreInt32(&(p.cancelled), math.MinInt32)
}

func (p *sliceSubscription) isCancelled() bool {
	return atomic.LoadInt32(&(p.cancelled)) != 0
}

func (p *sliceSubscription) slowPath(n int) {
	l := len(p.values)
	for n > 0 {
		next := int(atomic.AddInt32(&(p.cursor), 1))
		if next > l {
			return
		}
		v := p.values[next-1]
		if next == l {
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

func newSliceSubscription(s rs.Subscriber, values []interface{}) *sliceSubscription {
	return &sliceSubscription{
		actual: s,
		values: values,
	}
}

type sliceFlux struct {
	slice []interface{}
}

func (p *sliceFlux) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	if len(p.slice) < 1 {
		s.OnComplete()
		return
	}
	subscription := newSliceSubscription(s, p.slice)
	s.OnSubscribe(subscription)
}

func newSliceFlux(values []interface{}) *sliceFlux {
	return &sliceFlux{
		slice: values,
	}
}
