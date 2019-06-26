package mono

import (
	"context"
	"errors"
	"fmt"

	"github.com/jjeffcaii/reactor-go"
)

type defaultMonoProcessor struct {
	gen          func(Sink)
	hooks        *rs.Hooks
	pubScheduler rs.Scheduler
	subScheduler rs.Scheduler
	r            interface{}
	e            error
	sig          rs.Signal
	done         chan struct{}
	transforms   []rs.FnTransform
}

func (p *defaultMonoProcessor) Map(transform rs.FnTransform) Mono {
	if transform != nil {
		p.transforms = append(p.transforms, transform)
	}
	return p
}

func (p *defaultMonoProcessor) MapInt(transform func(int) interface{}) Mono {
	p.Map(func(i interface{}) interface{} {
		return transform(i.(int))
	})
	return p
}

func (p *defaultMonoProcessor) MapString(transform func(string) interface{}) Mono {
	p.Map(func(i interface{}) interface{} {
		return transform(i.(string))
	})
	return p
}

func (p *defaultMonoProcessor) Dispose() {
	panic("implement me")
}

func (p *defaultMonoProcessor) IsDisposed() bool {
	panic("implement me")
}

func (p *defaultMonoProcessor) Cancel() {
	if p.sig != rs.SignalDefault {
		return
	}
	p.sig = rs.SignalCancel
	close(p.done)
}

func (p *defaultMonoProcessor) Request(n int) {
}

func (p *defaultMonoProcessor) OnSubscribe(s rs.Subscription) {
	p.hooks.OnSubscribe(s)
}

func (p *defaultMonoProcessor) OnNext(s rs.Subscription, v interface{}) {
	p.hooks.OnNext(s, v)
}

func (p *defaultMonoProcessor) OnComplete() {
	p.hooks.OnComplete()
}

func (p *defaultMonoProcessor) OnError(err error) {
	p.hooks.OnError(err)
}

func (p *defaultMonoProcessor) Success(v interface{}) {
	if p.sig != rs.SignalDefault {
		return
	}
	for i, l := 0, len(p.transforms); i < l; i++ {
		v = p.transforms[i](v)
	}
	p.r = v
	p.sig = rs.SignalComplete
	close(p.done)
}

func (p *defaultMonoProcessor) Error(e error) {
	if p.sig != rs.SignalDefault {
		return
	}
	p.e = e
	p.sig = rs.SignalError
	close(p.done)
}

func (p *defaultMonoProcessor) SubscribeOn(s rs.Scheduler) Mono {
	p.subScheduler = s
	return p
}

func (p *defaultMonoProcessor) PublishOn(s rs.Scheduler) Mono {
	p.pubScheduler = s
	return p
}

func (p *defaultMonoProcessor) Subscribe(ctx context.Context, opts ...rs.OpSubscriber) rs.Disposable {
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
			rs.ReturnHooks(p.hooks)
			p.hooks = nil
		}()
		p.OnSubscribe(p)
		<-p.done
		switch p.sig {
		case rs.SignalComplete:
			p.OnNext(p, p.r)
		case rs.SignalError:
			p.OnError(p.e)
		case rs.SignalCancel:
			p.hooks.OnCancel()
		}
	})
	return p
}

func New(fn func(sink Sink)) Mono {
	return &defaultMonoProcessor{
		gen:          fn,
		done:         make(chan struct{}),
		sig:          rs.SignalDefault,
		pubScheduler: rs.Immediate(),
		subScheduler: rs.Immediate(),
		hooks:        rs.BorrowHooks(),
	}
}
