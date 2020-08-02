package flux

import (
	"context"
	"math"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

type Signal interface {
	Value() (Any, bool)
	Type() reactor.SignalType
}

type immutableSignal struct {
	v Any
	t reactor.SignalType
}

func (p *immutableSignal) Value() (v Any, ok bool) {
	ok = p.v != nil
	if ok {
		v = p.v
	}
	return
}

func (p *immutableSignal) Type() reactor.SignalType {
	return p.t
}

type fluxSwitchOnFirst struct {
	source reactor.RawPublisher
	t      FnSwitchOnFirst
}

func (p *fluxSwitchOnFirst) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	inner := newSwitchOnFirstInner(s, p.t)
	p.source.SubscribeWith(ctx, inner)
}

type switchOnFirstInner struct {
	inner       reactor.Subscriber
	outer       reactor.Subscriber
	transformer FnSwitchOnFirst
	s           reactor.Subscription
	first       Any
	stat        int32
}

func (p *switchOnFirstInner) SubscribeWith(ctx context.Context, actual reactor.Subscriber) {
	if !atomic.CompareAndSwapInt32(&p.stat, math.MinInt32, 0) {
		panic(errSubscribeOnce)
	}
	p.inner = actual
	actual.OnSubscribe(p)
}

func (p *switchOnFirstInner) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
	if p.first != nil {
		p.drain()
		if n < reactor.RequestInfinite {
			n--
			if n > 0 {
				p.s.Request(n)
			}
			return
		}
	}
	p.s.Request(n)
}

func (p *switchOnFirstInner) drain() {
	var first Any
	first, p.first = p.first, nil
	if first == nil {
		return
	}
	p.inner.OnNext(first)
}

func (p *switchOnFirstInner) Cancel() {
	if atomic.CompareAndSwapInt32(&p.stat, 0, statCancel) {
		p.s.Cancel()
	}
}

func (p *switchOnFirstInner) OnComplete() {
	if atomic.CompareAndSwapInt32(&p.stat, 0, statComplete) {
		p.inner.OnComplete()
	}
}

func (p *switchOnFirstInner) OnError(e error) {
	if atomic.CompareAndSwapInt32(&p.stat, 0, statError) {
		p.inner.OnError(e)
		return
	}
	hooks.Global().OnErrorDrop(e)
}

func (p *switchOnFirstInner) OnNext(v Any) {
	i := p.inner
	if i == nil {
		sig := &immutableSignal{
			v: v,
			t: reactor.SignalTypeDefault,
		}
		result := p.transformer(sig, wrap(p))
		p.first = v
		// TODO: which context?
		result.SubscribeWith(context.Background(), p.outer)
		return
	}
	i.OnNext(v)
}

func (p *switchOnFirstInner) OnSubscribe(s reactor.Subscription) {
	p.s = s
	s.Request(1)
}

type switchOnFirstInnerSubscriber struct {
	parent *switchOnFirstInner
	inner  reactor.Subscriber
}

func (p *switchOnFirstInnerSubscriber) OnComplete() {
	if atomic.LoadInt32(&p.parent.stat) == 0 {
		p.parent.Cancel()
	}
	p.inner.OnComplete()
}

func (p *switchOnFirstInnerSubscriber) OnError(e error) {
	if atomic.LoadInt32(&p.parent.stat) == 0 {
		p.parent.Cancel()
	}
	p.inner.OnError(e)
}

func (p *switchOnFirstInnerSubscriber) OnNext(t Any) {
	p.inner.OnNext(t)
}

func (p *switchOnFirstInnerSubscriber) OnSubscribe(s reactor.Subscription) {
	p.inner.OnSubscribe(s)
}

func newFluxSwitchOnFirst(source reactor.RawPublisher, transformer FnSwitchOnFirst) *fluxSwitchOnFirst {
	return &fluxSwitchOnFirst{
		source: source,
		t:      transformer,
	}
}

func newSwitchOnFirstInner(outer reactor.Subscriber, transformer FnSwitchOnFirst) *switchOnFirstInner {
	ret := &switchOnFirstInner{
		transformer: transformer,
		stat:        math.MinInt32,
	}
	ret.outer = newSwitchOnFirstInnerSubscriber(ret, outer)
	return ret
}

func newSwitchOnFirstInnerSubscriber(parent *switchOnFirstInner, inner reactor.Subscriber) *switchOnFirstInnerSubscriber {
	return &switchOnFirstInnerSubscriber{
		parent: parent,
		inner:  inner,
	}
}
