package mono

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
)

type Sink interface {
	Success(interface{})
	Error(error)
}

type defaultSink struct {
	s    rs.Subscriber
	stat int32
}

func (s *defaultSink) Success(v interface{}) {
	if v != nil {
		s.Next(v)
	}
	s.Complete()
}

func (s *defaultSink) Request(n int) {
	if n < 1 {
		panic(fmt.Errorf("negative request %d", n))
	}
}

func (s *defaultSink) Cancel() {
	atomic.StoreInt32(&(s.stat), math.MinInt32)
}

func (s *defaultSink) Complete() {
	if atomic.CompareAndSwapInt32(&(s.stat), 0, statComplete) {
		s.s.OnComplete()
	}
}

func (s *defaultSink) Error(err error) {
	if atomic.CompareAndSwapInt32(&(s.stat), 0, statError) {
		s.s.OnError(err)
	}
}

func (s *defaultSink) Next(v interface{}) {
	s.s.OnNext(s, v)
}

type monoCreate struct {
	*baseMono
	sinker func(context.Context, Sink)
}

func (m *monoCreate) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (m *monoCreate) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	sink := &defaultSink{
		s: s,
	}
	s.OnSubscribe(sink)
	m.sinker(ctx, sink)
}

func Create(gen func(context.Context, Sink)) Mono {
	m := &monoCreate{
		sinker: gen,
	}
	m.baseMono = &baseMono{
		child: m,
	}
	return m
}
