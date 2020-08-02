package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

type monoScheduleOn struct {
	source reactor.RawPublisher
	sc     scheduler.Scheduler
}

func (m *monoScheduleOn) SubscribeWith(ctx context.Context, s reactor.Subscriber) {
	actual := internal.NewCoreSubscriber(ctx, s)
	m.sc.Worker().Do(func() {
		m.source.SubscribeWith(ctx, actual)
	})
}

func newMonoScheduleOn(s reactor.RawPublisher, sc scheduler.Scheduler) *monoScheduleOn {
	return &monoScheduleOn{
		source: s,
		sc:     sc,
	}
}
