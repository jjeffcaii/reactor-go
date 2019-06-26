package rs

import (
	"context"
	"errors"
	"fmt"
)

var errWrongSignal = errors.New("wrong signal")
var errIllegalRequestN = errors.New("illegal requestN")

type fluxProcessor struct {
	n            int
	gen          func(context.Context, FluxSink)
	q            *queue
	e            error
	hooks        *Hooks
	sig          Signal
	pubScheduler Scheduler
	subScheduler Scheduler
	chain        []interface{}
}

func (p *fluxProcessor) Filter(fn FnFilter) Flux {
	if fn != nil {
		p.chain = append(p.chain, fn)
	}
	return p
}

func (p *fluxProcessor) Map(fn FnTransform) Flux {
	if fn != nil {
		p.chain = append(p.chain, fn)
	}
	return p
}

func (p *fluxProcessor) Dispose() {
	panic("implement me")
}

func (p *fluxProcessor) IsDisposed() bool {
	panic("implement me")
}

func (p *fluxProcessor) SubscribeOn(s Scheduler) Flux {
	p.subScheduler = s
	return p
}

func (p *fluxProcessor) PublishOn(s Scheduler) Flux {
	p.pubScheduler = s
	return p
}

func (p *fluxProcessor) Request(n int) {
	if n < 1 {
		panic(errIllegalRequestN)
	}
	if n > RequestInfinite {
		n = RequestInfinite
	}
	p.q.Request(int32(n))
}

func (p *fluxProcessor) Cancel() {
	p.sig = SignalCancel
	_ = p.q.Close()
}

func (p *fluxProcessor) Next(v interface{}) (err error) {
	if p.sig != SignalDefault {
		err = errWrongSignal
		return
	}
	for i, l := 0, len(p.chain); i < l; i++ {
		fn := p.chain[i]
		if t, ok := fn.(FnTransform); ok {
			v = t(v)
			continue
		}
		if !fn.(FnFilter)(v) {
			return
		}
	}
	return p.q.Push(v)
}

func (p *fluxProcessor) Error(e error) {
	p.e = e
	p.sig = SignalError
	_ = p.q.Close()
}

func (p *fluxProcessor) Complete() {
	p.sig = SignalComplete
	_ = p.q.Close()
}

func (p *fluxProcessor) Subscribe(ctx context.Context, opts ...OpSubscriber) Disposable {
	for _, it := range opts {
		it(p.hooks)
	}
	ctx, cancel := context.WithCancel(ctx)
	if p.gen != nil {
		p.pubScheduler.Do(ctx, func(ctx context.Context) {
			defer func() {
				e := recover()
				if e == nil {
					return
				}
				switch v := e.(type) {
				case error:
					p.Error(v)
				case string:
					p.Error(errors.New(v))
				default:
					p.Error(fmt.Errorf("%v", v))
				}
			}()
			p.gen(ctx, p)
		})
	}
	// bind request N
	p.q.HandleRequest(func(n int32) {
		p.hooks.OnRequest(int(n))
	})
	p.subScheduler.Do(ctx, func(ctx context.Context) {
		defer func() {
			p.hooks.OnFinally(p.sig)
			ReturnHooks(p.hooks)
			p.hooks = nil
		}()
		p.OnSubscribe(p)
		for {
			v, ok := p.q.Poll()
			if !ok {
				break
			}
			p.OnNext(p, v)
		}
		switch p.sig {
		case SignalComplete:
			p.OnComplete()
		case SignalError:
			p.OnError(p.e)
		case SignalCancel:
			cancel()
			p.hooks.OnCancel()
		}
	})
	return p
}

func (p *fluxProcessor) OnSubscribe(s Subscription) {
	p.hooks.OnSubscribe(s)
}

func (p *fluxProcessor) OnNext(s Subscription, v interface{}) {
	p.hooks.OnNext(s, v)
}

func (p *fluxProcessor) OnComplete() {
	p.hooks.OnComplete()
}

func (p *fluxProcessor) OnError(err error) {
	p.hooks.OnError(err)
}

func NewFlux(fn func(ctx context.Context, producer FluxSink)) Flux {
	return &fluxProcessor{
		hooks:        BorrowHooks(),
		gen:          fn,
		q:            newQueue(defaultQueueSize, RequestInfinite),
		pubScheduler: Elastic(),
		subScheduler: Immediate(),
	}
}
