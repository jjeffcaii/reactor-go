package mono

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type monoPeek[T any] struct {
	source          reactor.RawPublisher[T]
	onSubscribeCall reactor.FnOnSubscribe
	onNextCall      reactor.FnOnNext[T]
	onErrorCall     reactor.FnOnError
	onCompleteCall  reactor.FnOnComplete
	onRequestCall   reactor.FnOnRequest
	onCancelCall    reactor.FnOnCancel
}

func newMonoPeek[T any](source reactor.RawPublisher[T], first monoPeekOption[T], others ...monoPeekOption[T]) *monoPeek[T] {
	m := &monoPeek[T]{
		source: source,
	}
	first(m)
	for _, value := range others {
		value(m)
	}
	return m
}

func (p *monoPeek[T]) SubscribeWith(ctx context.Context, s reactor.Subscriber[T]) {
	p.source.SubscribeWith(ctx, newPeekSubscriber(p, s))
}

type peekSubscriber[T any] struct {
	actual    reactor.Subscriber[T]
	parent    *monoPeek[T]
	s         reactor.Subscription
	stat      int32
	cancelled int32
}

func newPeekSubscriber[T any](parent *monoPeek[T], actual reactor.Subscriber[T]) *peekSubscriber[T] {
	return &peekSubscriber[T]{
		parent: parent,
		actual: actual,
	}
}

func (p *peekSubscriber[T]) Request(ctx context.Context, n int) {
	if call := p.parent.onRequestCall; call != nil {
		call(ctx, n)
	}
	p.s.Request(ctx, n)
}

func (p *peekSubscriber[T]) Cancel(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&p.cancelled, 0, 1) {
		return
	}
	if call := p.parent.onCancelCall; call != nil {
		call(ctx)
	}
	p.s.Cancel(ctx)
}

func (p *peekSubscriber[T]) OnComplete(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statComplete) {
		return
	}
	if call := p.parent.onCompleteCall; call != nil {
		call(ctx)
	}
	p.actual.OnComplete(ctx)
}

func (p *peekSubscriber[T]) OnError(ctx context.Context, err error) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		if isCancelled := atomic.LoadInt32(&p.stat) == statCancel && errors.Is(err, reactor.ErrSubscribeCancelled); !isCancelled {
			return
		}
	}

	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			p.actual.OnError(ctx, errors.WithStack(e))
		} else {
			p.actual.OnError(ctx, errors.Errorf("%v", rec))
		}
	}()

	if call := p.parent.onErrorCall; call != nil {
		call(ctx, err)
	}
	p.actual.OnError(ctx, err)
}

func (p *peekSubscriber[T]) OnNext(ctx context.Context, v T) {
	if atomic.LoadInt32(&p.stat) != 0 {
		return
	}
	if call := p.parent.onNextCall; call != nil {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if e, ok := rec.(error); ok {
				p.OnError(ctx, errors.WithStack(e))
			} else {
				p.OnError(ctx, errors.Errorf("%v", rec))
			}
		}()
		if err := call(ctx, v); err != nil {
			defer p.OnError(ctx, err)
		}
	}
	p.actual.OnNext(ctx, v)
}

func (p *peekSubscriber[T]) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	if p.s != nil {
		panic(internal.ErrCallOnSubscribeDuplicated)
	}
	p.s = s
	if call := p.parent.onSubscribeCall; call != nil {
		call(ctx, p)
	}
	p.actual.OnSubscribe(ctx, p)
}

type monoPeekOption[T any] func(*monoPeek[T])

func peekNext[T any](fn reactor.FnOnNext[T]) monoPeekOption[T] {
	return func(peek *monoPeek[T]) {
		peek.onNextCall = fn
	}
}

func peekComplete[T any](fn reactor.FnOnComplete) monoPeekOption[T] {
	return func(peek *monoPeek[T]) {
		peek.onCompleteCall = fn
	}
}

func peekCancel[T any](fn reactor.FnOnCancel) monoPeekOption[T] {
	return func(peek *monoPeek[T]) {
		peek.onCancelCall = fn
	}
}

func peekError[T any](fn reactor.FnOnError) monoPeekOption[T] {
	return func(peek *monoPeek[T]) {
		peek.onErrorCall = fn
	}
}

func peekSubscribe[T any](fn reactor.FnOnSubscribe) monoPeekOption[T] {
	return func(peek *monoPeek[T]) {
		peek.onSubscribeCall = fn
	}
}

func peekRequest[T any](fn reactor.FnOnRequest) monoPeekOption[T] {
	return func(peek *monoPeek[T]) {
		peek.onRequestCall = fn
	}
}
