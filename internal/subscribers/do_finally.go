package subscribers

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

var _finallySubscriberPool = sync.Pool{
	New: func() interface{} {
		return new(DoFinallySubscriber)
	},
}

type DoFinallySubscriber struct {
	actual    reactor.Subscriber
	onFinally reactor.FnOnFinally
	s         reactor.Subscription
	done      int32
}

func BorrowDoFinallySubscriber(actual reactor.Subscriber, onFinally reactor.FnOnFinally) *DoFinallySubscriber {
	sub := _finallySubscriberPool.Get().(*DoFinallySubscriber)
	atomic.StoreInt32(&sub.done, 0)
	sub.actual = actual
	sub.onFinally = onFinally
	return sub
}

func ReturnDoFinallySubscriber(sub *DoFinallySubscriber) {
	sub.actual = nil
	sub.onFinally = nil
	sub.s = nil
	atomic.StoreInt32(&sub.done, 1)
	_finallySubscriberPool.Put(sub)
}

func (d *DoFinallySubscriber) Request(n int) {
	if atomic.LoadInt32(&d.done) == 0 {
		d.s.Request(n)
	}
}

func (d *DoFinallySubscriber) Cancel() {
	d.s.Cancel()
	d.runFinally(reactor.SignalTypeCancel)
}

func (d *DoFinallySubscriber) OnError(err error) {
	d.actual.OnError(err)
	if reactor.IsCancelledError(err) {
		d.runFinally(reactor.SignalTypeCancel)
	} else {
		d.runFinally(reactor.SignalTypeError)
	}
}

func (d *DoFinallySubscriber) OnNext(v reactor.Any) {
	d.actual.OnNext(v)
}

func (d *DoFinallySubscriber) OnSubscribe(ctx context.Context, s reactor.Subscription) {
	select {
	case <-ctx.Done():
		d.OnError(reactor.ErrSubscribeCancelled)
	default:
		d.s = s
		d.actual.OnSubscribe(ctx, d)
	}
}

func (d *DoFinallySubscriber) OnComplete() {
	d.actual.OnComplete()
	d.runFinally(reactor.SignalTypeComplete)
}

func (d *DoFinallySubscriber) runFinally(sig reactor.SignalType) {
	defer ReturnDoFinallySubscriber(d)
	if atomic.AddInt32(&d.done, 1) == 1 {
		d.onFinally(sig)
	}
}
