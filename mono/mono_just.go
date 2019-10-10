package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"go.uber.org/atomic"
)

type justSubscriber struct {
	s      rs.Subscriber
	parent *monoJust
	n      *atomic.Int32
}

func (j *justSubscriber) Request(n int) {
	if n < 1 {
		panic(rs.ErrNegativeRequest)
	}
	if !j.n.CAS(0, statComplete) {
		return
	}
	if j.parent.value == nil {
		j.s.OnComplete()
		return
	}
	defer func() {
		if err := internal.TryRecoverError(recover()); err != nil {
			j.s.OnError(err)
		} else {
			j.s.OnComplete()
		}
	}()
	j.s.OnNext(j.parent.value)
}

func (j *justSubscriber) Cancel() {
	j.n.CAS(0, statCancel)
}

type monoJust struct {
	value interface{}
}

func (m *monoJust) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	s.OnSubscribe(&justSubscriber{
		s:      internal.NewCoreSubscriber(ctx, s),
		parent: m,
		n:      atomic.NewInt32(0),
	})
}

func newMonoJust(v interface{}) *monoJust {
	return &monoJust{
		value: v,
	}
}
