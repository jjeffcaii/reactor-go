package flux

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type sliceSubscription struct {
	s         rs.Subscriber
	values    []interface{}
	cursor    int
	cancelled int32
}

func (p *sliceSubscription) Request(n int) {
	if n == rs.RequestInfinite {
		p.fastPath()
	} else {
		p.slowPath()
	}
}

func (p *sliceSubscription) Cancel() {
	atomic.StoreInt32(&(p.cancelled), math.MinInt32)
}

func (p *sliceSubscription) isCancelled() bool {
	return atomic.LoadInt32(&(p.cancelled)) != 0
}

func (p *sliceSubscription) slowPath() {
	panic("todo")
}

func (p *sliceSubscription) fastPath() {
	for i, l := p.cursor, len(p.values); i < l; i++ {
		if p.isCancelled() {
			return
		}
		v := p.values[i]
		if v == nil {
			p.s.OnError(fmt.Errorf("the %dth slice element was null", i))
			return
		}
		p.s.OnNext(p, v)
	}
	if p.isCancelled() {
		return
	}
	p.s.OnComplete()
}

func newSliceSubscription(s rs.Subscriber, values []interface{}) *sliceSubscription {
	return &sliceSubscription{
		s:      s,
		values: values,
	}
}

func newSliceFlux(values []interface{}) *sliceFlux {
	return &sliceFlux{
		slice: values,
	}
}

type sliceFlux struct {
	slice []interface{}
}

func (p *sliceFlux) Subscribe(ctx context.Context, s rs.Subscriber) {
	if len(p.slice) < 1 {
		s.OnComplete()
		return
	}
	subscription := newSliceSubscription(s, p.slice)
	s.OnSubscribe(subscription)
}

func (p *sliceFlux) Filter(filter rs.Predicate) Flux {
	return newFluxFilter(p, filter)
}

func (p *sliceFlux) Map(t rs.Transformer) Flux {
	return newFluxMap(p, t)
}

func (p *sliceFlux) SubscribeOn(sc scheduler.Scheduler) Flux {
	return newFluxSubscribeOn(p, sc)
}
