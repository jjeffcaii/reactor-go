package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"go.uber.org/atomic"
)

type peekSubscriber struct {
	actual rs.Subscriber
	parent *monoPeek
	s      rs.Subscription
	stat   *atomic.Int32
}

func (p *peekSubscriber) Request(n int) {
	p.s.Request(n)
}

func (p *peekSubscriber) Cancel() {
	defer p.s.Cancel()
	for _, fn := range p.parent.onCancelCall {
		fn()
	}
}

func (p *peekSubscriber) OnComplete() {
	if !p.stat.CAS(0, statComplete) {
		return
	}
	defer p.actual.OnComplete()
	for _, fn := range p.parent.onCompleteCall {
		fn()
	}
}

func (p *peekSubscriber) OnError(err error) {
	if !p.stat.CAS(0, statError) {
		return
	}
	defer p.actual.OnError(err)
	for _, fn := range p.parent.onErrorCall {
		fn(err)
	}
}

func (p *peekSubscriber) OnNext(v interface{}) {
	if p.stat.Load() != 0 {
		return
	}
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			p.OnError(err)
		}
	}()
	for _, fn := range p.parent.onNextCall {
		fn(v)
	}
	p.actual.OnNext(v)
}

func (p *peekSubscriber) OnSubscribe(s rs.Subscription) {
	if p.s != nil {
		panic(internal.ErrCallOnSubscribeDuplicated)
	}
	p.s = s
	defer p.actual.OnSubscribe(p)
	for _, fn := range p.parent.onSubscribeCall {
		fn(p)
	}
}

type monoPeek struct {
	source          rs.RawPublisher
	onSubscribeCall []rs.FnOnSubscribe
	onNextCall      []rs.FnOnNext
	onErrorCall     []rs.FnOnError
	onCompleteCall  []rs.FnOnComplete
	onCancelCall    []rs.FnOnCancel
}

func (p *monoPeek) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	p.source.SubscribeWith(ctx, internal.NewCoreSubscriber(ctx, &peekSubscriber{
		parent: p,
		actual: internal.ExtractRawSubscriber(actual),
		stat:   atomic.NewInt32(0),
	}))
}

func newMonoPeek(source rs.RawPublisher, first monoPeekOption, others ...monoPeekOption) (mp *monoPeek) {
	if m, ok := source.(*monoPeek); ok {
		mp = m
	} else {
		mp = &monoPeek{
			source: source,
		}
	}
	first(mp)
	for _, value := range others {
		value(mp)
	}
	return
}

type monoPeekOption func(*monoPeek)

func peekNext(fn rs.FnOnNext) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onNextCall = append(peek.onNextCall, fn)
	}
}

func peekComplete(fn rs.FnOnComplete) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onCompleteCall = append(peek.onCompleteCall, fn)
	}
}

func peekCancel(fn rs.FnOnCancel) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onCancelCall = append(peek.onCancelCall, fn)
	}
}

func peekError(fn rs.FnOnError) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onErrorCall = append(peek.onErrorCall, fn)
	}
}

func peekSubscribe(fn rs.FnOnSubscribe) monoPeekOption {
	return func(peek *monoPeek) {
		peek.onSubscribeCall = append(peek.onSubscribeCall, fn)
	}
}
