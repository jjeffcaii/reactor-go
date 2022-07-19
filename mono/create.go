package mono

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
)

type Sink interface {
	Success(Any)
	Error(error)
}

type monoCreate struct {
	sinker func(context.Context, Sink)
}

type sink struct {
	actual reactor.Subscriber
	stat   int32
}

func newMonoCreate(gen func(context.Context, Sink)) monoCreate {
	return monoCreate{
		sinker: func(ctx context.Context, sink Sink) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if e, ok := rec.(error); ok {
					sink.Error(errors.WithStack(e))
				} else {
					sink.Error(errors.Errorf("%v", rec))
				}
			}()

			select {
			case <-ctx.Done():
				sink.Error(reactor.NewContextError(ctx.Err()))
			default:
				gen(ctx, sink)
			}
		},
	}
}

func (s *sink) Success(v Any) {
	if atomic.LoadInt32(&s.stat) != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	if v != nil {
		s.Next(v)
	}
	s.Complete()
}

func (s *sink) Request(_ int) {
	// ignore
}

func (s *sink) Cancel() {
	if !atomic.CompareAndSwapInt32(&s.stat, 0, statCancel) {
		return
	}
	s.actual.OnError(reactor.ErrSubscribeCancelled)
}

func (s *sink) Complete() {
	if !atomic.CompareAndSwapInt32(&s.stat, 0, statComplete) {
		return
	}
	s.actual.OnComplete()
}

func (s *sink) Error(err error) {
	if !atomic.CompareAndSwapInt32(&s.stat, 0, statError) {
		hooks.Global().OnErrorDrop(err)
		return
	}
	s.actual.OnError(err)
}

func (s *sink) Next(v Any) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		if e, ok := rec.(error); ok {
			s.Error(errors.WithStack(e))
		} else {
			s.Error(errors.Errorf("%v", rec))
		}
	}()
	s.actual.OnNext(v)
}

func (m monoCreate) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	var sk sink
	sk.actual = s
	s.OnSubscribe(ctx, &sk)
	m.sinker(ctx, &sk)
}
