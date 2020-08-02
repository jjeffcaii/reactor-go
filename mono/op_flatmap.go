package mono

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

const (
	statCancel   = -1
	statError    = -2
	statComplete = 2
)

type innerFlatMapSubscriber struct {
	parent *flatMapSubscriber
}

func (in *innerFlatMapSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&in.parent.stat, 0, statError) {
		in.parent.actual.OnError(err)
	}
}

func (in *innerFlatMapSubscriber) OnNext(v Any) {
	if atomic.LoadInt32(&in.parent.stat) != 0 {
		return
	}
	in.parent.actual.OnNext(v)
	in.OnComplete()
}

func (in *innerFlatMapSubscriber) OnSubscribe(s reactor.Subscription) {
	s.Request(reactor.RequestInfinite)
}

func (in *innerFlatMapSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&in.parent.stat, 0, statComplete) {
		in.parent.actual.OnComplete()
	}
}

type flatMapSubscriber struct {
	actual reactor.Subscriber
	mapper flatMapper
	stat   int32
	ctx    context.Context
}

func (p *flatMapSubscriber) Request(n int) {
}

func (p *flatMapSubscriber) Cancel() {
	atomic.CompareAndSwapInt32(&p.stat, 0, statCancel)
}

func (p *flatMapSubscriber) OnComplete() {
	if atomic.LoadInt32(&p.stat) == statComplete {
		p.actual.OnComplete()
	}
}

func (p *flatMapSubscriber) OnError(err error) {
	if atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		p.actual.OnError(err)
	}
}

func (p *flatMapSubscriber) OnNext(v Any) {
	if atomic.LoadInt32(&p.stat) != 0 {
		return
	}
	m := p.mapper(v)
	inner := &innerFlatMapSubscriber{
		parent: p,
	}
	m.SubscribeWith(p.ctx, inner)
}

func (p *flatMapSubscriber) OnSubscribe(s reactor.Subscription) {
	s.Request(reactor.RequestInfinite)
}

func newFlatMapSubscriber(ctx context.Context, actual reactor.Subscriber, mapper flatMapper) *flatMapSubscriber {
	return &flatMapSubscriber{
		ctx:    ctx,
		actual: actual,
		mapper: mapper,
	}
}

type monoFlatMap struct {
	source reactor.RawPublisher
	mapper flatMapper
}

func (m *monoFlatMap) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	s := newFlatMapSubscriber(ctx, actual, m.mapper)
	actual.OnSubscribe(s)
	m.source.SubscribeWith(ctx, s)
}

func newMonoFlatMap(source reactor.RawPublisher, mapper flatMapper) *monoFlatMap {
	return &monoFlatMap{
		source: source,
		mapper: mapper,
	}
}
