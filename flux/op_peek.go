package flux

import (
	"context"

	rs "github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"go.uber.org/atomic"
)

type peekSubscriber struct {
	parent *fluxPeek
	actual rs.Subscriber
	s      rs.Subscription
	stat   *atomic.Int32
}

func (p *peekSubscriber) Request(n int) {
	for _, fn := range p.parent.onRequestCall {
		fn(n)
	}
	p.s.Request(n)
}

func (p *peekSubscriber) Cancel() {
	for _, fn := range p.parent.onCancelCall {
		fn()
	}
	p.s.Cancel()
}

func (p *peekSubscriber) OnComplete() {
	if !p.stat.CAS(0, statComplete) {
		return
	}
	for _, fn := range p.parent.onCompleteCall {
		fn()
	}
	p.actual.OnComplete()
}

func (p *peekSubscriber) OnError(e error) {
	if !p.stat.CAS(0, statError) {
		return
	}
	for _, fn := range p.parent.onErrorCall {
		fn(e)
	}
	p.actual.OnError(e)
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
	for _, fn := range p.parent.onSubscribeCall {
		fn(p)
	}
	p.actual.OnSubscribe(p)
}

func newPeekSubscriber(peek *fluxPeek, actual rs.Subscriber) *peekSubscriber {
	return &peekSubscriber{
		parent: peek,
		actual: actual,
		stat:   atomic.NewInt32(0),
	}
}

type fluxPeek struct {
	source          rs.RawPublisher
	onSubscribeCall []rs.FnOnSubscribe
	onNextCall      []rs.FnOnNext
	onErrorCall     []rs.FnOnError
	onCompleteCall  []rs.FnOnComplete
	onRequestCall   []rs.FnOnRequest
	onCancelCall    []rs.FnOnCancel
}

func (p *fluxPeek) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	actual = internal.ExtractRawSubscriber(actual)
	actual = internal.NewCoreSubscriber(ctx, newPeekSubscriber(p, actual))
	p.source.SubscribeWith(ctx, actual)
}

func newFluxPeek(source rs.RawPublisher, options ...fluxPeekOption) (fp *fluxPeek) {
	// try merge
	if f, ok := source.(*fluxPeek); ok {
		fp = f
	} else {
		fp = &fluxPeek{
			source: source,
		}
	}
	for i := range options {
		options[i](fp)
	}
	return fp
}

type fluxPeekOption func(*fluxPeek)

func peekNext(fn rs.FnOnNext) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onNextCall = append(peek.onNextCall, fn)
	}
}

func peekComplete(fn rs.FnOnComplete) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onCompleteCall = append(peek.onCompleteCall, fn)
	}
}

func peekCancel(fn rs.FnOnCancel) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onCancelCall = append(peek.onCancelCall, fn)
	}
}
func peekRequest(fn rs.FnOnRequest) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onRequestCall = append(peek.onRequestCall, fn)
	}
}

func peekError(fn rs.FnOnError) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onErrorCall = append(peek.onErrorCall, fn)
	}
}

func peekSubscribe(fn rs.FnOnSubscribe) fluxPeekOption {
	return func(peek *fluxPeek) {
		peek.onSubscribeCall = append(peek.onSubscribeCall, fn)
	}
}
