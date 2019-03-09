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
	gen          func(context.Context, Producer)
	q            *queue
	e            error
	hooks        *hooks
	sig          Signal
	pubScheduler Scheduler
	subScheduler Scheduler
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
	if p.sig == SignalDefault {
		err = p.q.Push(v)
	} else {
		err = errWrongSignal
	}
	return
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
		p.hooks.OnRequest(ctx, int(n))
	})
	p.subScheduler.Do(ctx, func(ctx context.Context) {
		defer func() {
			p.hooks.OnFinally(ctx, p.sig)
			returnHooks(p.hooks)
			p.hooks = nil
		}()
		p.OnSubscribe(ctx, p)
		for {
			v, ok := p.q.Poll()
			if !ok {
				break
			}
			p.OnNext(ctx, p, v)
		}
		switch p.sig {
		case SignalComplete:
			p.OnComplete(ctx)
		case SignalError:
			p.OnError(ctx, p.e)
		case SignalCancel:
			cancel()
			p.hooks.OnCancel(ctx)
		}
	})
	return p
}

func (p *fluxProcessor) OnSubscribe(ctx context.Context, s Subscription) {
	p.hooks.OnSubscribe(ctx, s)
}

func (p *fluxProcessor) OnNext(ctx context.Context, s Subscription, v interface{}) {
	p.hooks.OnNext(ctx, s, v)
}

func (p *fluxProcessor) OnComplete(ctx context.Context) {
	p.hooks.OnComplete(ctx)
}

func (p *fluxProcessor) OnError(ctx context.Context, err error) {
	p.hooks.OnError(ctx, err)
}

func NewFlux(fn func(ctx context.Context, producer Producer)) Flux {
	return &fluxProcessor{
		hooks:        borrowHooks(),
		gen:          fn,
		q:            newQueue(defaultQueueSize, RequestInfinite),
		pubScheduler: Elastic(),
		subScheduler: Immediate(),
	}
}
