package rs

import (
	"context"
	"errors"
	"fmt"
)

type defaultMonoProcessor struct {
	gen          func(MonoSink)
	hooks        *Hooks
	pubScheduler Scheduler
	subScheduler Scheduler
	r            interface{}
	e            error
	sig          Signal
	done         chan struct{}
	transforms   []FnTransform
}

func (p *defaultMonoProcessor) Map(transform FnTransform) Mono {
	if transform != nil {
		p.transforms = append(p.transforms, transform)
	}
	return p
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

func (p *defaultMonoProcessor) OnSubscribe(s Subscription) {
	p.hooks.OnSubscribe(s)
}

func (p *defaultMonoProcessor) OnNext(s Subscription, v interface{}) {
	p.hooks.OnNext(s, v)
}

func (p *defaultMonoProcessor) OnComplete() {
	p.hooks.OnComplete()
}

func (p *defaultMonoProcessor) OnError(err error) {
	p.hooks.OnError(err)
}

func (p *defaultMonoProcessor) Success(v interface{}) {
	if p.sig != SignalDefault {
		return
	}
	for i, l := 0, len(p.transforms); i < l; i++ {
		v = p.transforms[i](v)
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
			p.hooks.OnFinally(p.sig)
			ReturnHooks(p.hooks)
			p.hooks = nil
		}()
		p.OnSubscribe(p)
		<-p.done
		switch p.sig {
		case SignalComplete:
			p.OnNext(p, p.r)
		case SignalError:
			p.OnError(p.e)
		case SignalCancel:
			p.hooks.OnCancel()
		}
	})
	return p
}

func NewMono(fn func(sink MonoSink)) Mono {
	return &defaultMonoProcessor{
		gen:          fn,
		done:         make(chan struct{}),
		sig:          SignalDefault,
		pubScheduler: Immediate(),
		subScheduler: Immediate(),
		hooks:        BorrowHooks(),
	}
}
