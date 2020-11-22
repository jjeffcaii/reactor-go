package subscribers

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

var _doFinallySubscriberPool = sync.Pool{
	New: func() interface{} {
		return new(DoFinallySubscriber)
	},
}

var (
	_ reactor.Subscriber = (*DoFinallySubscriber)(nil)
	_ reactor.Disposable = (*DoFinallySubscriber)(nil)
)

type DoFinallySubscriber struct {
	actual    reactor.Subscriber
	onFinally reactor.FnOnFinally
	s         reactor.Subscription
	done      int32
}

func (d *DoFinallySubscriber) Dispose() {
	d.actual = nil
	d.onFinally = nil
	d.s = nil
	_doFinallySubscriberPool.Put(d)
}

func NewDoFinallySubscriber(actual reactor.Subscriber, onFinally reactor.FnOnFinally) *DoFinallySubscriber {
	s := _doFinallySubscriberPool.Get().(*DoFinallySubscriber)
	s.actual = actual
	s.onFinally = onFinally
	atomic.StoreInt32(&s.done, 0)
	return s
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
	if atomic.AddInt32(&d.done, 1) == 1 {
		onFinally := d.onFinally
		d.Dispose()
		onFinally(sig)
	}
}
