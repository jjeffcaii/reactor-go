package subscribers

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
)

type DoFinallySubscriber struct {
	actual    reactor.Subscriber
	onFinally reactor.FnOnFinally
	done      chan struct{}
	s         reactor.Subscription
}

func NewDoFinallySubscriber(actual reactor.Subscriber, onFinally reactor.FnOnFinally) *DoFinallySubscriber {
	return &DoFinallySubscriber{
		actual:    actual,
		onFinally: onFinally,
		done:      make(chan struct{}),
	}
}

func (p *DoFinallySubscriber) Request(n int) {
	p.s.Request(n)
}

func (p *DoFinallySubscriber) Cancel() {
	p.s.Cancel()
	p.runFinally(reactor.SignalTypeCancel)
}

func (p *DoFinallySubscriber) OnError(err error) {
	p.actual.OnError(err)
	if reactor.IsCancelledError(err) {
		p.runFinally(reactor.SignalTypeCancel)
	} else {
		p.runFinally(reactor.SignalTypeError)
	}
}

func (p *DoFinallySubscriber) OnNext(v reactor.Any) {
	p.actual.OnNext(v)
}

func (p *DoFinallySubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	select {
	case <-ctx.Done():
		p.OnError(reactor.ErrSubscribeCancelled)
	default:
		p.s = s
		p.actual.OnSubscribe(ctx, p)
	}
}

func (p *DoFinallySubscriber) OnComplete() {
	p.actual.OnComplete()
	p.runFinally(reactor.SignalTypeComplete)
}

func (p *DoFinallySubscriber) runFinally(sig reactor.SignalType) {
	select {
	case <-p.done:
	default:
		if !internal.SafeCloseDone(p.done) {
			return
		}
		p.onFinally(sig)
	}
}
