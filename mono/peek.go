package mono

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type monoPeek struct {
	source          reactor.RawPublisher
	onSubscribeCall reactor.FnOnSubscribe
	onNextCall      reactor.FnOnNext
	onErrorCall     reactor.FnOnError
	onCompleteCall  reactor.FnOnComplete
	onRequestCall   reactor.FnOnRequest
	onCancelCall    reactor.FnOnCancel
}

type peekSubscriber struct {
	actual    reactor.Subscriber
	parent    *monoPeek
	s         reactor.Subscription
	stat      int32
	cancelled int32
}

func newMonoPeek(source reactor.RawPublisher, first monoPeekOption, others ...monoPeekOption) *monoPeek {
	m := &monoPeek{
		source: source,
	}
	first(m)
	for _, value := range others {
		value(m)
	}
	return m
}

func newPeekSubscriber(parent *monoPeek, actual reactor.Subscriber) *peekSubscriber {
	return &peekSubscriber{
		parent: parent,
		actual: actual,
	}
}

func (p *peekSubscriber) Request(n int) {
	if call := p.parent.onRequestCall; call != nil {
		call(n)
	}
	p.s.Request(n)
}

func (p *peekSubscriber) Cancel() {
	if !atomic.CompareAndSwapInt32(&p.cancelled, 0, 1) {
		return
	}
	if call := p.parent.onCancelCall; call != nil {
		call()
	}
	p.s.Cancel()
}

func (p *peekSubscriber) OnComplete() {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statComplete) {
		return
	}
	if call := p.parent.onCompleteCall; call != nil {
		call()
	}
	p.actual.OnComplete()
}

func (p *peekSubscriber) OnError(err error) {
	if !atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		if isCancelled := atomic.LoadInt32(&p.stat) == statCancel && err == reactor.ErrSubscribeCancelled; !isCancelled {
			return
		}
	}

	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			p.actual.OnError(errors.WithStack(e))
		} else {
			p.actual.OnError(errors.Errorf("%v", rec))
		}
	}()

	if call := p.parent.onErrorCall; call != nil {
		call(err)
	}
	p.actual.OnError(err)
}

func (p *peekSubscriber) OnNext(v Any) {
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
				p.OnError(errors.WithStack(e))
			} else {
				p.OnError(errors.Errorf("%v", rec))
			}
		}()
		if err := call(v); err != nil {
			defer p.OnError(err)
		}
	}
	p.actual.OnNext(v)
}

func (p *peekSubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	if p.s != nil {
		panic(internal.ErrCallOnSubscribeDuplicated)
	}
	p.s = s
	if call := p.parent.onSubscribeCall; call != nil {
		call(ctx, p)
	}
	p.actual.OnSubscribe(ctx, p)
}

func (p *monoPeek) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	p.source.SubscribeWith(ctx, newPeekSubscriber(p, s))
}

type monoPeekOption func(*monoPeek)

func peekNext(fn reactor.FnOnNext) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onNextCall = fn
	}
}

func peekComplete(fn reactor.FnOnComplete) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onCompleteCall = fn
	}
}

func peekCancel(fn reactor.FnOnCancel) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onCancelCall = fn
	}
}

func peekError(fn reactor.FnOnError) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onErrorCall = fn
	}
}

func peekSubscribe(fn reactor.FnOnSubscribe) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onSubscribeCall = fn
	}
}

func peekRequest(fn reactor.FnOnRequest) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onRequestCall = fn
	}
}
