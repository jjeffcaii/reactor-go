package flux

import (
	"context"
	"sync"
	"sync/atomic"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type fluxRange struct {
	begin, end int64
}

func (r fluxRange) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	raw := internal.ExtractRawSubscriber(s)
	su := newRangeSubscription(raw, r.begin, r.end)
	internal.NewCoreSubscriber(ctx, raw).OnSubscribe(su)
}

type rangeSubscription struct {
	cursor, begin, end int64
	actual             rs.Subscriber
	flags              uint8
	locker             sync.Mutex
}

func (p *rangeSubscription) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
	}
	p.locker.Lock()
	if p.flags&flagFast != 0 {
		p.locker.Unlock()
		return
	}
	if n < rs.RequestInfinite {
		p.locker.Unlock()
		p.slowPath(n)
		return
	}
	p.flags |= flagFast
	p.locker.Unlock()
	p.fastPath()
}

func (p *rangeSubscription) Cancel() {
	p.locker.Lock()
	p.flags |= flagCancel
	p.locker.Unlock()
}

func (p *rangeSubscription) slowPath(n int) {
	for n > 0 {
		next := p.begin + atomic.AddInt64(&(p.cursor), 1)
		if next > p.end {
			return
		}
		v := next - 1
		if next == p.end {
			p.actual.OnNext(v)
			p.actual.OnComplete()
			return
		}
		n--
		p.actual.OnNext(v)
	}
}

func (p *rangeSubscription) fastPath() {
	for i := p.begin; i < p.end; i++ {
		if p.isCancelled() {
			return
		}
		v := i
		p.actual.OnNext(v)
	}
	if p.isCancelled() {
		return
	}
	p.actual.OnComplete()
}

func (p *rangeSubscription) isCancelled() (cancelled bool) {
	p.locker.Lock()
	cancelled = p.flags&flagCancel != 0
	p.locker.Unlock()
	return
}

func newRangeSubscription(actual rs.Subscriber, begin, end int64) *rangeSubscription {
	return &rangeSubscription{
		actual: actual,
		begin:  begin,
		end:    end,
	}
}

func newFluxRange(begin, end int64) fluxRange {
	return fluxRange{
		begin: begin,
		end:   end,
	}
}
