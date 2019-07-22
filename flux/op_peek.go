package flux

import (
	"context"
	"sync/atomic"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type peekSubscriber struct {
	parent *fluxPeek
	actual rs.Subscriber
	s      rs.Subscription
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

func (p *peekSubscriber) OnNext(v interface{}) {
	if atomic.LoadInt32(&(p.stat)) != 0 {
		return
	}
	if call := p.parent.onNextCall; call != nil {
		defer func() {
			if err := internal.TryRecoverError(recover()); err != nil {
				p.OnError(err)
			}
		}()
		call(v)
	}
	p.actual.OnNext(v)
}

func (p *peekSubscriber) OnSubscribe(s rs.Subscription) {
	if p.s != nil {
		panic(internal.ErrCallOnSubscribeDuplicated)
	}
	p.s = s
	if call := p.parent.onSubscribeCall; call != nil {
		call(p)
	}
	p.actual.OnSubscribe(p)
}

func newPeekSubscriber(peek *fluxPeek, actual rs.Subscriber) *peekSubscriber {
	return &peekSubscriber{
		parent: peek,
		actual: actual,
	}
}

type fluxPeek struct {
	source          rs.RawPublisher
	onSubscribeCall rs.FnOnSubscribe
	onNextCall      rs.FnOnNext
	onErrorCall     rs.FnOnError
	onCompleteCall  rs.FnOnComplete
	onRequestCall   rs.FnOnRequest
	onCancelCall    rs.FnOnCancel
}

func (p *fluxPeek) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, newPeekSubscriber(p, actual))
	p.source.SubscribeWith(ctx, actual)
}

func newFluxPeek(source rs.RawPublisher, options ...fluxPeekOption) *fluxPeek {
	ret := &fluxPeek{
		source: source,
	}
	for i := range options {
		options[i](ret)
	}
	return ret
}

type fluxPeekOption func(*fluxPeek)

func peekNext(fn rs.FnOnNext) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onNextCall = fn
	}
}

func peekComplete(fn rs.FnOnComplete) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onCompleteCall = fn
	}
}

func peekCancel(fn rs.FnOnCancel) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onCancelCall = fn
	}
}
func peekRequest(fn rs.FnOnRequest) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onRequestCall = fn
	}
}

func peekError(fn rs.FnOnError) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onErrorCall = fn
	}
}

func peekSubscribe(fn rs.FnOnSubscribe) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onSubscribeCall = fn
	}
}
