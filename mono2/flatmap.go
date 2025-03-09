package mono

import (
	"context"
	stdErrors "errors"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/jjeffcaii/reactor-go"
)

var errNilFlatMapResult = stdErrors.New("the FlatMap result is nil")

type FlatMapper[A, B any] func(ctx context.Context, value A) Mono[B]

type flatMapStat int32

const (
	_ flatMapStat = iota
	flatMapStatInnerReady
	flatMapStatInnerComplete
	flatMapStatError
	flatMapStatComplete
)

type innerFlatMapSubscriber[A, B any] struct {
	parent *flatMapSubscriber[A, B]
}

func (in *innerFlatMapSubscriber[A, B]) OnError(ctx context.Context, err error) {
	if atomic.CompareAndSwapInt32(&in.parent.stat, int32(flatMapStatInnerReady), int32(flatMapStatError)) {
		in.parent.actual.OnError(ctx, err)
	}
}

func (in *innerFlatMapSubscriber[A, B]) OnNext(ctx context.Context, v B) {
	if atomic.LoadInt32(&in.parent.stat) != int32(flatMapStatInnerReady) {
		// TODO: hooks
		//hooks.Global().OnNextDrop(v)
		return
	}
	in.parent.actual.OnNext(ctx, v)
	in.OnComplete(ctx)
}

func (in *innerFlatMapSubscriber[A, B]) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	s.Request(ctx, reactor.RequestInfinite)
}

func (in *innerFlatMapSubscriber[A, B]) OnComplete(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&in.parent.stat, int32(flatMapStatInnerReady), int32(flatMapStatInnerComplete)) {
		in.parent.actual.OnComplete(ctx)
	}
}

type flatMapSubscriber[A, B any] struct {
	actual reactor.Subscriber[B]
	mapper FlatMapper[A, B]
	stat   int32
	ctx    context.Context
}

func newFlatMapSubscriber[A, B any](actual reactor.Subscriber[B], mapper FlatMapper[A, B]) *flatMapSubscriber[A, B] {
	return &flatMapSubscriber[A, B]{
		actual: actual,
		mapper: mapper,
	}
}

func (p *flatMapSubscriber[A, B]) Request(ctx context.Context, n int) {
}

func (p *flatMapSubscriber[A, B]) Cancel(ctx context.Context) {
	atomic.CompareAndSwapInt32(&p.stat, 0, statCancel)
}

func (p *flatMapSubscriber[A, B]) OnComplete(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&p.stat, 0, int32(flatMapStatComplete)) {
		p.actual.OnComplete(ctx)
	}
}

func (p *flatMapSubscriber[A, B]) OnError(ctx context.Context, err error) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, int32(flatMapStatError)) {
		// TODO: hooks
		//hooks.Global().OnErrorDrop(err)
		return
	}
	p.actual.OnError(ctx, err)
}

func (p *flatMapSubscriber[A, B]) OnNext(ctx context.Context, v A) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, int32(flatMapStatInnerReady)) {
		// TODO: hooks
		//hooks.Global().OnNextDrop(v)
		return
	}
	nextMono, err := p.computeNextMono(ctx, v)
	if err != nil {
		p.actual.OnError(ctx, err)
		return
	}
	inner := &innerFlatMapSubscriber[A, B]{
		parent: p,
	}
	nextMono.SubscribeWith(p.ctx, inner)
}

func (p *flatMapSubscriber[A, B]) computeNextMono(ctx context.Context, v A) (next Mono[B], err error) {
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
	next = p.mapper(ctx, v)
	if next == nil {
		err = errNilFlatMapResult
	}
	return
}

func (p *flatMapSubscriber[A, B]) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	p.ctx = ctx
	s.Request(ctx, reactor.RequestInfinite)
}

type monoFlatMap[A, B any] struct {
	source reactor.RawPublisher[A]
	mapper FlatMapper[A, B]
}

func newMonoFlatMap[A, B any](source reactor.RawPublisher[A], mapper FlatMapper[A, B]) *monoFlatMap[A, B] {
	return &monoFlatMap[A, B]{
		source: source,
		mapper: mapper,
	}
}

func (m *monoFlatMap[A, B]) SubscribeWith(ctx context.Context, actual reactor.Subscriber[B]) {
	s := newFlatMapSubscriber(actual, m.mapper)
	actual.OnSubscribe(ctx, s)
	m.source.SubscribeWith(ctx, s)
}
