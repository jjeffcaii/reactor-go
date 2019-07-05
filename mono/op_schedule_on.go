package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type monoScheduleOn struct {
	*baseMono
	source Mono
	sc     scheduler.Scheduler
}

func (m *monoScheduleOn) Subscribe(ctx context.Context, options ...rs.SubscriberOption) {
	m.SubscribeWith(ctx, rs.NewSubscriber(options...))
}

func (m *monoScheduleOn) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	w := m.sc.Worker()
	w.Do(func() {
		defer func() {
			_ = w.Close()
		}()
		m.source.SubscribeWith(ctx, s)
	})
}

func newMonoScheduleOn(s Mono, sc scheduler.Scheduler) Mono {
	m := &monoScheduleOn{
		source: s,
		sc:     sc,
	}
	m.baseMono = &baseMono{
		child: m,
	}
	return m
}
