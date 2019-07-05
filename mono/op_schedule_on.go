package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type monoScheduleOn struct {
	source Mono
	sc     scheduler.Scheduler
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

func newMonoScheduleOn(s Mono, sc scheduler.Scheduler) *monoScheduleOn {
	return &monoScheduleOn{
		source: s,
		sc:     sc,
	}
}
