package flux

import (
	"context"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type peekSubscriber struct {
	parent *fluxPeek
	actual reactor.Subscriber
	s      reactor.Subscription
	stat   int32
}

func (p *peekSubscriber) Request(n int) {
	if call := p.parent.onRequestCall; call != nil {
		call(n)
	}
	p.s.Request(n)
}

func (p *peekSubscriber) Cancel() {
	if call := p.parent.onCancelCall; call != nil {
		call()
	}
	p.s.Cancel()
}

func (p *peekSubscriber) OnComplete() {
	if !atomic.CompareAndSwapInt32(&(p.stat), 0, statComplete) {
		return
	}
	if call := p.parent.onCompleteCall; call != nil {
		call()
	}
	p.actual.OnComplete()
}

func (p *peekSubscriber) OnError(e error) {
	if !atomic.CompareAndSwapInt32(&(p.stat), 0, statError) {
		return
	}
	if call := p.parent.onErrorCall; call != nil {
		call(e)
	}
	p.actual.OnError(e)
}

func (p *peekSubscriber) OnNext(v Any) {
	if atomic.LoadInt32(&(p.stat)) != 0 {
		return
	}
	if call := p.parent.onNextCall; call != nil {
		if err := call(v); err != nil {
			defer p.OnError(err)
		}
	}
	p.actual.OnNext(v)
}

func (p *peekSubscriber) OnSubscribe(s reactor.Subscription) {
	if p.s != nil {
		panic(internal.ErrCallOnSubscribeDuplicated)
	}
	p.s = s
	if call := p.parent.onSubscribeCall; call != nil {
		call(p)
	}
	p.actual.OnSubscribe(p)
}

func newPeekSubscriber(peek *fluxPeek, actual reactor.Subscriber) *peekSubscriber {
	return &peekSubscriber{
		parent: peek,
		actual: actual,
	}
}

type fluxPeek struct {
	source          reactor.RawPublisher
	onSubscribeCall reactor.FnOnSubscribe
	onNextCall      reactor.FnOnNext
	onErrorCall     reactor.FnOnError
	onCompleteCall  reactor.FnOnComplete
	onRequestCall   reactor.FnOnRequest
	onCancelCall    reactor.FnOnCancel
}

func (p *fluxPeek) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, newPeekSubscriber(p, actual))
	p.source.SubscribeWith(ctx, actual)
}

func newFluxPeek(source reactor.RawPublisher, options ...fluxPeekOption) *fluxPeek {
	ret := &fluxPeek{
		source: source,
	}
	for i := range options {
		options[i](ret)
	}
	return ret
}

type fluxPeekOption func(*fluxPeek)

func peekNext(fn reactor.FnOnNext) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onNextCall = fn
	}
}

func peekComplete(fn reactor.FnOnComplete) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onCompleteCall = fn
	}
}

func peekCancel(fn reactor.FnOnCancel) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onCancelCall = fn
	}
}
func peekRequest(fn reactor.FnOnRequest) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onRequestCall = fn
	}
}

func peekError(fn reactor.FnOnError) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onErrorCall = fn
	}
}

func peekSubscribe(fn reactor.FnOnSubscribe) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onSubscribeCall = fn
	}
}
