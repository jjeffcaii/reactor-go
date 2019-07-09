package mono

import (
	"context"
	"errors"
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
	if !atomic.CompareAndSwapInt32(&(j.n), 0, statComplete) {
		return
	}
	if j.v == nil {
		j.s.OnComplete()
		return
	}
	defer func() {
		re := recover()
		if re == nil {
			j.s.OnComplete()
			return
		}
		switch v := re.(type) {
		case error:
			j.s.OnError(v)
		case string:
			j.s.OnError(errors.New(v))
		default:
			j.s.OnError(fmt.Errorf("%s", v))
		}
	}()
	j.s.OnNext(j.v)
}

func (j *justSubscriber) Cancel() {
	atomic.CompareAndSwapInt32(&(j.n), 0, statCancel)
}

type monoJust struct {
	value interface{}
}

func (m *monoJust) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (m *monoJust) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	s.OnSubscribe(&justSubscriber{
		s: s,
		v: m.value,
	})
}

func newMonoJust(v interface{}) *monoJust {
	return &monoJust{
		value: v,
	}
}
