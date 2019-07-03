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
	if !atomic.CompareAndSwapInt32(&(s.stat), 0, 1) {
		return
	}
}

func (s *defaultSink) Cancel() {
	atomic.StoreInt32(&(s.stat), math.MinInt32)
}

func (s *defaultSink) Complete() {
	if atomic.CompareAndSwapInt32(&(s.stat), 1, 2) {
		s.s.OnComplete()
	}
}

func (s *defaultSink) Error(err error) {
	if atomic.CompareAndSwapInt32(&(s.stat), 0, 1) {
		s.s.OnError(err)
	}
}

func (s *defaultSink) Next(v interface{}) {
	s.s.OnNext(s, v)
}

type monoCreate struct {
	sinker func(context.Context, Sink)
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

func (m monoCreate) Subscribe(ctx context.Context, s rs.Subscriber) {
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