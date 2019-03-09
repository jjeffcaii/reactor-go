package rs

import (
	"context"
	"errors"
	"fmt"
)

type defaultMonoProcessor struct {
	gen          func(MonoProducer)
	hooks        *hooks
	pubScheduler Scheduler
	subScheduler Scheduler
	r            interface{}
	e            error
	sig          Signal
	done         chan struct{}
}

func (p *defaultMonoProcessor) Dispose() {
	panic("implement me")
}

func (p *defaultMonoProcessor) IsDisposed() bool {
	panic("implement me")
}

func (p *defaultMonoProcessor) Cancel() {
	if p.sig != SignalDefault {
		return
	}
	p.sig = SignalCancel
	close(p.done)
}

func (p *defaultMonoProcessor) Request(n int) {
}

func (p *defaultMonoProcessor) OnSubscribe(ctx context.Context, s Subscription) {
	p.hooks.OnSubscribe(ctx, s)
}

func (p *defaultMonoProcessor) OnNext(ctx context.Context, s Subscription, v interface{}) {
	p.hooks.OnNext(ctx, s, v)
}

func (p *defaultMonoProcessor) OnComplete(ctx context.Context) {
	p.hooks.OnComplete(ctx)
}

func (p *defaultMonoProcessor) OnError(ctx context.Context, err error) {
	p.hooks.OnError(ctx, err)
}

func (p *defaultMonoProcessor) Success(v interface{}) {
	if p.sig != SignalDefault {
		return
	}
	p.r = v
	p.sig = SignalComplete
	close(p.done)
}

func (p *defaultMonoProcessor) Error(e error) {
	if p.sig != SignalDefault {
		return
	}
	p.e = e
	p.sig = SignalError
	close(p.done)
}

func (p *defaultMonoProcessor) SubscribeOn(s Scheduler) Mono {
	p.subScheduler = s
	return p
}

func (p *defaultMonoProcessor) PublishOn(s Scheduler) Mono {
	p.pubScheduler = s
	return p
}

func (p *defaultMonoProcessor) Subscribe(ctx context.Context, opts ...OpSubscriber) Disposable {
	for _, fn := range opts {
		fn(p.hooks)
	}
	p.pubScheduler.Do(ctx, func(ctx context.Context) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			switch v := rec.(type) {
			case error:
				p.Error(v)
			case string:
				p.Error(errors.New(v))
			default:
				p.Error(fmt.Errorf("%v", v))
			}
		}()
		p.gen(p)
	})
	p.subScheduler.Do(ctx, func(ctx context.Context) {
		defer func() {
			p.hooks.OnFinally(ctx, p.sig)
			returnHooks(p.hooks)
			p.hooks = nil
		}()
		p.OnSubscribe(ctx, p)
		<-p.done
		switch p.sig {
		case SignalComplete:
			p.OnNext(ctx, p, p.r)
		case SignalError:
			p.OnError(ctx, p.e)
		case SignalCancel:
			p.hooks.OnCancel(ctx)
		}
	})
	return p
}

func NewMono(fn func(sink MonoProducer)) Mono {
	return &defaultMonoProcessor{
		gen:          fn,
		done:         make(chan struct{}),
		sig:          SignalDefault,
		pubScheduler: Immediate(),
		subScheduler: Immediate(),
		hooks:        borrowHooks(),
	}
}
