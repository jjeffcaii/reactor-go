package mono

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
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
	s.Next(v)
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
	sinker func(context.Context, Sink)
}

func (m monoCreate) DoFinally(fn rs.FnOnFinally) Mono {
	return newMonoDoFinally(m, fn)
}

func (m monoCreate) DoOnNext(fn rs.FnOnNext) Mono {
	return newMonoPeek(m, peekNext(fn))
}

func (m monoCreate) Block(ctx context.Context) (interface{}, error) {
	return toBlock(ctx, m)
}

func (m monoCreate) FlatMap(f flatMapper) Mono {
	return newMonoFlatMap(m, f)
}

func (m monoCreate) SubscribeOn(sc scheduler.Scheduler) Mono {
	return newMonoScheduleOn(m, sc)
}

func (m monoCreate) Filter(f rs.Predicate) Mono {
	return newMonoFilter(m, f)
}

func (m monoCreate) Map(t rs.Transformer) Mono {
	return newMonoMap(m, t)
}

func (m monoCreate) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeRaw(ctx, rs.NewSubscriber(options...))
}

func (m monoCreate) SubscribeRaw(ctx context.Context, s rs.Subscriber) {
	sink := &defaultSink{
		s: s,
	}
	s.OnSubscribe(sink)
	m.sinker(ctx, sink)
}

func Create(gen func(context.Context, Sink)) Mono {
	return monoCreate{
		sinker: gen,
	}
}
