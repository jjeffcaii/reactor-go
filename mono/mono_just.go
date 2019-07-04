package mono

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type justSubscriber struct {
	s rs.Subscriber
	v interface{}
	n int32
}

func (j *justSubscriber) Request(n int) {
	if n < 1 {
		panic(fmt.Errorf("negative request %d", n))
	}
	if !atomic.CompareAndSwapInt32(&(j.n), 0, 1) {
		return
	}
	defer j.s.OnComplete()
	j.s.OnNext(j, j.v)
}

func (j *justSubscriber) Cancel() {
	atomic.CompareAndSwapInt32(&(j.n), 0, -1)
}

type monoJust struct {
	value interface{}
}

func (m *monoJust) DoOnNext(fn rs.FnOnNext) Mono {
	return newMonoPeek(m, peekNext(fn))
}

func (m *monoJust) Block(ctx context.Context) (interface{}, error) {
	return toBlock(ctx, m)
}

func (m *monoJust) FlatMap(f flatMapper) Mono {
	return newMonoFlatMap(m, f)
}

func (m *monoJust) SubscribeOn(sc scheduler.Scheduler) Mono {
	return newMonoScheduleOn(m, sc)
}

func (m *monoJust) Filter(f rs.Predicate) Mono {
	return newMonoFilter(m, f)
}

func (m *monoJust) Map(t rs.Transformer) Mono {
	return newMonoMap(m, t)
}

func (m *monoJust) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeRaw(ctx, rs.NewSubscriber(options...))
}

func (m *monoJust) SubscribeRaw(ctx context.Context, s rs.Subscriber) {
	s.OnSubscribe(&justSubscriber{
		s: s,
		v: m.value,
	})
}

func Just(v interface{}) Mono {
	return &monoJust{
		value: v,
	}
}
