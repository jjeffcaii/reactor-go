package flux

import (
	"context"
	"sync"
	"sync/atomic"

	"golang.org/x/exp/constraints"

	"github.com/jjeffcaii/reactor-go"
)

func Range[T constraints.Signed](start T, count int) Flux[T] {
	if count < 1 {
		// TODO: impl empty
		return Just[T]()
	}
	return wrap(newFluxRange(start, start+T(count)))
}

type fluxRange[T constraints.Signed] struct {
	begin, end T
}

func newFluxRange[T constraints.Signed](begin, end T) fluxRange[T] {
	return fluxRange[T]{
		begin: begin,
		end:   end,
	}
}

func (r fluxRange[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	su := newRangeSubscription(s, r.begin, r.end)
	s.OnSubscribe(ctx, su)
}

type rangeSubscription[T constraints.Signed] struct {
	cursor     int32
	begin, end T
	actual     reactor.Subscriber[T]
	flags      uint8
	locker     sync.Mutex
}

func newRangeSubscription[T constraints.Signed](actual reactor.Subscriber[T], begin, end T) *rangeSubscription[T] {
	return &rangeSubscription[T]{
		actual: actual,
		begin:  begin,
		end:    end,
	}
}

func (p *rangeSubscription[T]) Request(ctx context.Context, n int) {
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
		p.slowPath(ctx, n)
		return
	}
	p.flags |= flagFast
	p.locker.Unlock()
	p.fastPath(ctx)
}

func (p *rangeSubscription[T]) Cancel(ctx context.Context) {
	p.locker.Lock()
	p.flags |= flagCancel
	p.locker.Unlock()
}

func (p *rangeSubscription[T]) slowPath(ctx context.Context, n int) {
	for n > 0 {
		next := p.begin + T(atomic.AddInt32(&p.cursor, 1))
		if next > p.end {
			return
		}
		v := next - 1
		if next == p.end {
			p.actual.OnNext(ctx, v)
			p.actual.OnComplete(ctx)
			return
		}
		n--
		p.actual.OnNext(ctx, v)
	}
}

func (p *rangeSubscription[T]) fastPath(ctx context.Context) {
	for i := p.begin; i < p.end; i++ {
		if p.isCancelled() {
			return
		}
		v := i
		p.actual.OnNext(ctx, v)
	}
	if p.isCancelled() {
		return
	}
	p.actual.OnComplete(ctx)
}

func (p *rangeSubscription[T]) isCancelled() (cancelled bool) {
	p.locker.Lock()
	cancelled = p.flags&flagCancel != 0
	p.locker.Unlock()
	return
}
