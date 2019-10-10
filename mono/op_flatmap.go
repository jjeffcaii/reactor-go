package mono

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type innerFlatMapSubscriber struct {
	parent *flatMapSubscriber
}

func (p *innerFlatMapSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&(p.parent.stat), 0, statError) {
		p.parent.actual.OnError(err)
	}
}

func (p *innerFlatMapSubscriber) OnNext(v interface{}) {
	if atomic.LoadInt32(&(p.parent.stat)) != 0 {
		return
	}
	p.parent.actual.OnNext(v)
	p.OnComplete()
}

func (p *innerFlatMapSubscriber) OnSubscribe(s rs.Subscription) {
	s.Request(rs.RequestInfinite)
}

func (p *innerFlatMapSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&(p.parent.stat), 0, statComplete) {
		p.parent.actual.OnComplete()
	}
}

type flatMapSubscriber struct {
	actual rs.Subscriber
	mapper flatMapper
	stat   int32
	ctx    context.Context
}

func (p *flatMapSubscriber) Request(n int) {
}

func (p *flatMapSubscriber) Cancel() {
	atomic.CompareAndSwapInt32(&(p.stat), 0, statCancel)
}

func (p *flatMapSubscriber) OnComplete() {
	if atomic.LoadInt32(&(p.stat)) == statComplete {
		p.actual.OnComplete()
	}
}

func (p *flatMapSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&(p.stat), 0, statError) {
		p.actual.OnError(err)
	}
}

func (p *flatMapSubscriber) OnNext(v interface{}) {
	if atomic.LoadInt32(&(p.stat)) != 0 {
		return
	}
	m := p.mapper(v)
	inner := &innerFlatMapSubscriber{
		parent: p,
	}
	m.SubscribeWith(p.ctx, inner)
}

func (p *flatMapSubscriber) OnSubscribe(s rs.Subscription) {
	s.Request(rs.RequestInfinite)
}

func newFlatMapSubscriber(ctx context.Context, actual rs.Subscriber, mapper flatMapper) *flatMapSubscriber {
	return &flatMapSubscriber{
		ctx:    ctx,
		actual: actual,
		mapper: mapper,
	}
}

type monoFlatMap struct {
	source rs.RawPublisher
	mapper flatMapper
}

func (m *monoFlatMap) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	s := newFlatMapSubscriber(ctx, actual, m.mapper)
	actual.OnSubscribe(s)
	m.source.SubscribeWith(ctx, s)
}

func newMonoFlatMap(source rs.RawPublisher, mapper flatMapper) *monoFlatMap {
	return &monoFlatMap{
		source: source,
		mapper: mapper,
	}
}
