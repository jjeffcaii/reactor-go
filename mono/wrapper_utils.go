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
	s := subscribers.BorrowBlockSubscriber()
	defer subscribers.ReturnBlockSubscriber(s)

	publisher.SubscribeWith(ctx, s)
	<-s.Done()

	if s.E != nil {
		return nil, s.E
	}
	return s.V, nil
}

func mustProcessor(publisher reactor.RawPublisher) rawProcessor {
	rp, ok := publisher.(rawProcessor)
	if !ok {
		panic(errNotProcessor)
	}
	return rp
}
