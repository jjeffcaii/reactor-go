package flux

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type peekSubscriber[T any] struct {
	parent *fluxPeek[T]
	actual reactor.Subscriber[T]
	s      reactor.Subscription
	stat   int32
}

func newPeekSubscriber[T any](peek *fluxPeek[T], actual reactor.Subscriber[T]) *peekSubscriber[T] {
	return &peekSubscriber[T]{
		parent: peek,
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

func (p *peekSubscriber[T]) OnError(ctx context.Context, e error) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		return
	}
	if call := p.parent.onErrorCall; call != nil {
		call(ctx, e)
	}
	p.actual.OnError(ctx, e)
}

func (p *peekSubscriber[T]) OnNext(ctx context.Context, v T) {
	if atomic.LoadInt32(&p.stat) != 0 {
		return
	}
	if call := p.parent.onNextCall; call != nil {
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

type fluxPeek[T any] struct {
	source          reactor.RawPublisher[T]
	onSubscribeCall reactor.FnOnSubscribe
	onNextCall      reactor.FnOnNext[T]
	onErrorCall     reactor.FnOnError
	onCompleteCall  reactor.FnOnComplete
	onRequestCall   reactor.FnOnRequest
	onCancelCall    reactor.FnOnCancel
}

func newFluxPeek[T any](source reactor.RawPublisher[T], options ...fluxPeekOption[T]) *fluxPeek[T] {
	ret := &fluxPeek[T]{
		source: source,
	}
	for i := range options {
		options[i](ret)
	}
	return ret
}

func (fp *fluxPeek[T]) SubscribeWith(ctx context.Context, actual reactor.Subscriber[T]) {
	fp.source.SubscribeWith(ctx, newPeekSubscriber(fp, actual))
}

type fluxPeekOption[T any] func(*fluxPeek[T])

func peekNext[T any](fn reactor.FnOnNext[T]) fluxPeekOption[T] {
	return func(peek *fluxPeek[T]) {
		peek.onNextCall = fn
	}
}

func peekComplete[T any](fn reactor.FnOnComplete) fluxPeekOption[T] {
	return func(peek *fluxPeek[T]) {
		peek.onCompleteCall = fn
	}
}

func peekCancel[T any](fn reactor.FnOnCancel) fluxPeekOption[T] {
	return func(peek *fluxPeek[T]) {
		peek.onCancelCall = fn
	}
}
func peekRequest[T any](fn reactor.FnOnRequest) fluxPeekOption[T] {
	return func(peek *fluxPeek[T]) {
		peek.onRequestCall = fn
	}
}

func peekError[T any](fn reactor.FnOnError) fluxPeekOption[T] {
	return func(peek *fluxPeek[T]) {
		peek.onErrorCall = fn
	}
}

func peekSubscribe[T any](fn reactor.FnOnSubscribe) fluxPeekOption[T] {
	return func(peek *fluxPeek[T]) {
		peek.onSubscribeCall = fn
	}
}
