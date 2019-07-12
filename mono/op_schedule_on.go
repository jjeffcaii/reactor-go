package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type monoScheduleOn struct {
	source Mono
	sc     scheduler.Scheduler
}

func (m *monoScheduleOn) SubscribeWith(ctx context.Context, s rs.Subscriber) {
	actual := internal.NewCoreSubscriber(ctx, s)
	w := m.sc.Worker()
	w.Do(func() {
		defer func() {
			_ = w.Close()
		}()
		m.source.SubscribeWith(ctx, actual)
	})
}

func newMonoScheduleOn(s Mono, sc scheduler.Scheduler) *monoScheduleOn {
	return &monoScheduleOn{
		source: s,
		sc:     sc,
	}
}
