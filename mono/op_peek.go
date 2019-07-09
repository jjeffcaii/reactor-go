package mono

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

type peekSubscriber struct {
	actual rs.Subscriber
	parent *monoPeek
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

func (p *peekSubscriber) OnError(err error) {
	if !atomic.CompareAndSwapInt32(&(p.stat), 0, statError) {
		return
	}
	if call := p.parent.onErrorCall; call != nil {
		call(err)
	}
	p.actual.OnError(err)
}

func (p *peekSubscriber) OnNext(v interface{}) {
	if atomic.LoadInt32(&(p.stat)) != 0 {
		return
	}
	if call := p.parent.onNextCall; call != nil {
		call(v)
	}
	p.actual.OnNext(v)
}

func (p *peekSubscriber) OnSubscribe(s rs.Subscription) {
	if p.s != nil {
		panic(errors.New("call lOnSubscribe duplicated"))
	}
	p.s = s
	if call := p.parent.onSubscribeCall; call != nil {
		call(s)
	}
	p.actual.OnSubscribe(p)
}

func newPeekSubscriber(parent *monoPeek, actual rs.Subscriber) *peekSubscriber {
	return &peekSubscriber{
		parent: parent,
		actual: actual,
	}
}

type monoPeek struct {
	source          Mono
	onSubscribeCall rs.FnOnSubscribe
	onNextCall      rs.FnOnNext
	onErrorCall     rs.FnOnError
	onCompleteCall  rs.FnOnComplete
	onRequestCall   rs.FnOnRequest
	onCancelCall    rs.FnOnCancel
}

func (p *monoPeek) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	p.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (p *monoPeek) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	p.source.SubscribeWith(ctx, newPeekSubscriber(p, s))
}

type monoPeekOption func(*monoPeek)

func peekNext(fn rs.FnOnNext) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onNextCall = fn
	}
}

func peekComplete(fn rs.FnOnComplete) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onCompleteCall = fn
	}
}

func peekCancel(fn rs.FnOnCancel) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onCancelCall = fn
	}
}
func peekRequest(fn rs.FnOnRequest) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onRequestCall = fn
	}
}

func peekError(fn rs.FnOnError) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onErrorCall = fn
	}
}

func peekSubscribe(fn rs.FnOnSubscribe) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onSubscribeCall = fn
	}
}

func newMonoPeek(source Mono, first monoPeekOption, others ...monoPeekOption) *monoPeek {
	m := &monoPeek{
		source: source,
	}
	first(m)
	for _, value := range others {
		value(m)
	}
	return m
}
