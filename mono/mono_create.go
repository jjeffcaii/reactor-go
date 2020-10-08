package mono

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
)

var _errRunSinkFailed = errors.New("execute creation func failed")

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

func newMonoCreate(gen func(context.Context, Sink)) *monoCreate {
	return &monoCreate{
		sinker: func(ctx context.Context, sink Sink) {
			defer func() {
				if e := recover(); e != nil {
					sink.Error(_errRunSinkFailed)
				}
			}()

			select {
			case <-ctx.Done():
				sink.Error(ctx.Err())
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

func (s *sink) Request(n int) {
	if n < 1 {
		panic(reactor.ErrNegativeRequest)
	}
}

func (s *sink) Cancel() {
	atomic.CompareAndSwapInt32(&s.stat, 0, statCancel)
}

func (s *sink) Complete() {
	if atomic.CompareAndSwapInt32(&s.stat, 0, statComplete) {
		s.actual.OnComplete()
	}
}

func (s *sink) Error(err error) {
	if atomic.CompareAndSwapInt32(&s.stat, 0, statError) {
		s.actual.OnError(err)
		return
	}
	hooks.Global().OnErrorDrop(err)
}

func (s *sink) Next(v Any) {
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			s.Error(err)
		}
	}()
	s.actual.OnNext(v)
}

func (m *monoCreate) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	sink := &sink{
		actual: s,
	}
	s.OnSubscribe(ctx, sink)
	m.sinker(ctx, sink)
}
