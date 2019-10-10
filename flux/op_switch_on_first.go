package flux

import (
	"context"
	"math"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"go.uber.org/atomic"
)

type Signal interface {
	Value() (interface{}, bool)
	Type() rs.SignalType
}

type immutableSignal struct {
	v interface{}
	t rs.SignalType
}

func (p *immutableSignal) Value() (v interface{}, ok bool) {
	ok = p.v != nil
	if ok {
		v = p.v
	}
	return
}

func (p *immutableSignal) Type() rs.SignalType {
	return p.t
}

type fluxSwitchOnFirst struct {
	source rs.RawPublisher
	t      FnSwitchOnFirst
}

func (p *fluxSwitchOnFirst) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	inner := newSwitchOnFirstInner(s, p.t)
	p.source.SubscribeWith(ctx, inner)
}

type switchOnFirstInner struct {
	inner       rs.Subscriber
	outer       rs.Subscriber
	transformer FnSwitchOnFirst
	s           rs.Subscription
	first       interface{}
	stat        *atomic.Int32
}

func (p *switchOnFirstInner) SubscribeWith(ctx context.Context, actual rs.Subscriber) {
	if !p.stat.CAS(math.MinInt32, 0) {
		panic(errSubscribeOnce)
	}
	p.inner = actual
	actual.OnSubscribe(p)
}

func (p *switchOnFirstInner) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
	}
	if p.first != nil {
		p.drain()
		if n < rs.RequestInfinite {
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
	var first interface{}
	first, p.first = p.first, nil
	if first == nil {
		return
	}
	p.inner.OnNext(first)
}

func (p *switchOnFirstInner) Cancel() {
	if p.stat.CAS(0, statCancel) {
		p.s.Cancel()
	}
}

func (p *switchOnFirstInner) OnComplete() {
	if p.stat.CAS(0, statComplete) {
		p.inner.OnComplete()
	}
}

func (p *switchOnFirstInner) OnError(e error) {
	if p.stat.CAS(0, statError) {
		p.inner.OnError(e)
		return
	}
	hooks.Global().OnErrorDrop(e)
}

func (p *switchOnFirstInner) OnNext(v interface{}) {
	i := p.inner
	if i == nil {
		sig := &immutableSignal{
			v: v,
			t: rs.SignalTypeDefault,
		}
		result := p.transformer(sig, wrap(p))
		p.first = v
		// TODO: which context?
		result.SubscribeWith(context.Background(), p.outer)
		return
	}
	i.OnNext(v)
}

func (p *switchOnFirstInner) OnSubscribe(s rs.Subscription) {
	p.s = s
	s.Request(1)
}

type switchOnFirstInnerSubscriber struct {
	parent *switchOnFirstInner
	inner  rs.Subscriber
}

func (p *switchOnFirstInnerSubscriber) OnComplete() {
	if p.parent.stat.Load() == 0 {
		p.parent.Cancel()
	}
	p.inner.OnComplete()
}

func (p *switchOnFirstInnerSubscriber) OnError(e error) {
	if p.parent.stat.Load() == 0 {
		p.parent.Cancel()
	}
	p.inner.OnError(e)
}

func (p *switchOnFirstInnerSubscriber) OnNext(t interface{}) {
	p.inner.OnNext(t)
}

func (p *switchOnFirstInnerSubscriber) OnSubscribe(s rs.Subscription) {
	p.inner.OnSubscribe(s)
}

func newFluxSwitchOnFirst(source rs.RawPublisher, transformer FnSwitchOnFirst) *fluxSwitchOnFirst {
	return &fluxSwitchOnFirst{
		source: source,
		t:      transformer,
	}
}

func newSwitchOnFirstInner(outer rs.Subscriber, transformer FnSwitchOnFirst) *switchOnFirstInner {
	ret := &switchOnFirstInner{
		transformer: transformer,
		stat:        atomic.NewInt32(math.MinInt32),
	}
	ret.outer = newSwitchOnFirstInnerSubscriber(ret, outer)
	return ret
}

func newSwitchOnFirstInnerSubscriber(parent *switchOnFirstInner, inner rs.Subscriber) *switchOnFirstInnerSubscriber {
	return &switchOnFirstInnerSubscriber{
		parent: parent,
		inner:  inner,
	}
}
