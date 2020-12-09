package mono

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/pkg/errors"
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

func (in *innerFlatMapSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	s.Request(reactor.RequestInfinite)
}

func (in *innerFlatMapSubscriber) OnComplete() {
	if atomic.CompareAndSwapInt32(&in.parent.stat, 0, statComplete) {
		in.parent.actual.OnComplete()
	}
}

type flatMapSubscriber struct {
	actual reactor.Subscriber
	mapper FlatMapper
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
	nextMono, err := p.computeNextMono(v)
	if err != nil {
		p.actual.OnError(err)
		return
	}
	inner := &innerFlatMapSubscriber{
		parent: p,
	}
	nextMono.SubscribeWith(p.ctx, inner)
}

func (p *flatMapSubscriber) computeNextMono(v Any) (next Mono, err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			err = errors.WithStack(e)
		} else {
			err = errors.Errorf("%v", rec)
		}
	}()
	next = p.mapper(v)
	if next == nil {
		err = errors.New("the FlatMap result is nil")
	}
	return
}

func (p *flatMapSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	p.ctx = ctx
	s.Request(reactor.RequestInfinite)
}

func newFlatMapSubscriber(actual reactor.Subscriber, mapper FlatMapper) *flatMapSubscriber {
	return &flatMapSubscriber{
		actual: actual,
		mapper: mapper,
	}
}

type monoFlatMap struct {
	source reactor.RawPublisher
	mapper FlatMapper
}

func (m *monoFlatMap) Parent() reactor.RawPublisher {
	return m.source
}

func (m *monoFlatMap) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	s := newFlatMapSubscriber(actual, m.mapper)
	actual.OnSubscribe(ctx, s)
	m.source.SubscribeWith(ctx, s)
}

func newMonoFlatMap(source reactor.RawPublisher, mapper FlatMapper) *monoFlatMap {
	return &monoFlatMap{
		source: source,
		mapper: mapper,
	}
}
