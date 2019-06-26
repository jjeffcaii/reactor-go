package flux

import (
	"context"
	"errors"
	"fmt"

	"github.com/jjeffcaii/reactor-go"
)

var errWrongSignal = errors.New("wrong signal")
var errIllegalRequestN = errors.New("illegal requestN")

type fluxProcessor struct {
	n            int
	gen          func(context.Context, Sink)
	q            *Queue
	e            error
	hooks        *rs.Hooks
	sig          rs.Signal
	pubScheduler rs.Scheduler
	subScheduler rs.Scheduler
	chain        []interface{}
}

func (p *fluxProcessor) Filter(fn rs.Predicate) Flux {
	if fn != nil {
		p.chain = append(p.chain, fn)
	}
	return p
}

func (p *fluxProcessor) Map(fn rs.FnTransform) Flux {
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

func (p *fluxProcessor) SubscribeOn(s rs.Scheduler) Flux {
	p.subScheduler = s
	return p
}

func (p *fluxProcessor) PublishOn(s rs.Scheduler) Flux {
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
	p.sig = rs.SignalCancel
	_ = p.q.Close()
}

func (p *fluxProcessor) Next(v interface{}) (err error) {
	if p.sig != rs.SignalDefault {
		err = errWrongSignal
		return
	}
	for i, l := 0, len(p.chain); i < l; i++ {
		fn := p.chain[i]
		if t, ok := fn.(rs.FnTransform); ok {
			v = t(v)
			continue
		}
		if !fn.(rs.Predicate)(v) {
			return
		}
	}
	return p.q.Push(v)
}

func (p *fluxProcessor) Error(e error) {
	p.e = e
	p.sig = rs.SignalError
	_ = p.q.Close()
}

func (p *fluxProcessor) Complete() {
	p.sig = rs.SignalComplete
	_ = p.q.Close()
}

func (p *fluxProcessor) Subscribe(ctx context.Context, opts ...rs.OpSubscriber) rs.Disposable {
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
			rs.ReturnHooks(p.hooks)
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
		case rs.SignalComplete:
			p.OnComplete()
		case rs.SignalError:
			p.OnError(p.e)
		case rs.SignalCancel:
			cancel()
			p.hooks.OnCancel()
		}
	})
	return p
}

func (p *fluxProcessor) OnSubscribe(s rs.Subscription) {
	p.hooks.OnSubscribe(s)
}

func (p *fluxProcessor) OnNext(s rs.Subscription, v interface{}) {
	p.hooks.OnNext(s, v)
}

func (p *fluxProcessor) OnComplete() {
	p.hooks.OnComplete()
}

func (p *fluxProcessor) OnError(err error) {
	p.hooks.OnError(err)
}

func New(fn func(ctx context.Context, producer Sink)) Flux {
	return &fluxProcessor{
		hooks:        rs.BorrowHooks(),
		gen:          fn,
		q:            NewQueue(DefaultQueueSize, RequestInfinite),
		pubScheduler: rs.Elastic(),
		subScheduler: rs.Immediate(),
	}
}
