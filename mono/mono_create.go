package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/hooks"
	"github.com/jjeffcaii/reactor-go/internal"
	"go.uber.org/atomic"
)

type Sink interface {
	Success(interface{})
	Error(error)
}

type defaultSink struct {
	actual rs.Subscriber
	stat   *atomic.Int32
}

func (s *defaultSink) Success(v interface{}) {
	if s.stat.Load() != 0 {
		hooks.Global().OnNextDrop(v)
		return
	}
	if v != nil {
		s.Next(v)
	}
	s.Complete()
}

func (s *defaultSink) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
	}
}

func (s *defaultSink) Cancel() {
	s.stat.CAS(0, statCancel)
}

func (s *defaultSink) Complete() {
	if s.stat.CAS(0, statComplete) {
		s.actual.OnComplete()
	}
}

func (s *defaultSink) Error(err error) {
	if s.stat.CAS(0, statError) {
		s.actual.OnError(err)
		return
	}
	hooks.Global().OnErrorDrop(err)
}

func (s *defaultSink) Next(v interface{}) {
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			s.Error(err)
		}
	}()
	s.actual.OnNext(v)
}

type monoCreate struct {
	sinker func(context.Context, Sink)
}

func (m *monoCreate) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	sink := &defaultSink{
		actual: s,
		stat:   atomic.NewInt32(0),
	}
	s.OnSubscribe(sink)
	m.sinker(ctx, sink)
}

func newMonoCreate(gen func(context.Context, Sink)) *monoCreate {
	return &monoCreate{
		sinker: func(i context.Context, sink Sink) {
			defer func() {
				if err := internal.TryRecoverError(recover()); err != nil {
					sink.Error(err)
				}
			}()
			gen(i, sink)
		},
	}
}
