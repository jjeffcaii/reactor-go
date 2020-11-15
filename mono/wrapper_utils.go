package mono

import (
	"context"

	"github.com/jjeffcaii/reactor-go"
	"github.com/jjeffcaii/reactor-go/internal/subscribers"
	"github.com/jjeffcaii/reactor-go/scheduler"
)

func IsSubscribeAsync(m Mono) bool {
	var publisher reactor.RawPublisher
	switch t := m.(type) {
	case wrapper:
		publisher = t.RawPublisher
	case *oneshotWrapper:
		publisher = t.RawPublisher
	default:
		return false
	}

	var sc scheduler.Scheduler
	switch pub := publisher.(type) {
	case monoScheduleOn:
		sc = pub.sc
	case *monoScheduleOn:
		sc = pub.sc
	default:
		return false
	}
	return scheduler.IsParallel(sc) || scheduler.IsElastic(sc) || scheduler.IsSingle(sc)
}

func block(ctx context.Context, publisher reactor.RawPublisher) (Any, error) {
	done := make(chan struct{})
	c := make(chan reactor.Item, 1)
	b := subscribers.NewBlockSubscriber(done, c)
	publisher.SubscribeWith(ctx, b)
	<-done
	defer close(c)

	select {
	case result := <-c:
		if result.E != nil {
			return nil, result.E
		}
		return result.V, nil
	default:
		return nil, nil
	}
}

func mustProcessor(publisher reactor.RawPublisher) *processor {
	pp, ok := publisher.(*processor)
	if !ok {
		panic(errNotProcessor)
	}
	return pp
}
