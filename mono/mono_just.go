package mono

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/jjeffcaii/reactor-go"
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

func (m *monoJust) Map(t rs.Transformer) Mono {
	return newMonoMap(m, t)
}

func (m *monoJust) Subscribe(ctx context.Context, s rs.Subscriber) rs.Disposable {
	s.OnSubscribe(&justSubscriber{
		s: s,
		v: m.value,
	})
	return nil
}

func Just(v interface{}) Mono {
	return &monoJust{
		value: v,
	}
}
